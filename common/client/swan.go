package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"go-swan-client/common/constants"
	"go-swan-client/common/utils"
	"go-swan-client/config"
	"go-swan-client/logs"
	"go-swan-client/model"
)

const GET_OFFLINEDEAL_LIMIT_DEFAULT = 50
const RESPONSE_STATUS_SUCCESS = "SUCCESS"

type TokenAccessInfo struct {
	ApiKey      string `json:"apikey"`
	AccessToken string `json:"access_token"`
}

type SwanClient struct {
	ApiUrl string
	ApiKey string
	Token  string
}

type GetOfflineDealResponse struct {
	Data   GetOfflineDealData `json:"data"`
	Status string             `json:"status"`
}

type GetOfflineDealData struct {
	Deal []model.OfflineDeal `json:"deal"`
}

type UpdateOfflineDealResponse struct {
	Data   UpdateOfflineDealData `json:"data"`
	Status string                `json:"status"`
}

type UpdateOfflineDealData struct {
	Deal    model.OfflineDeal `json:"deal"`
	Message string            `json:"message"`
}

func SwanGetClient() *SwanClient {
	mainConf := config.GetConfig().Main
	uri := mainConf.SwanApiUrl + "/user/api_keys/jwt"
	data := TokenAccessInfo{ApiKey: mainConf.SwanApiKey, AccessToken: mainConf.SwanAccessToken}
	response := HttpPostNoToken(uri, data)

	if strings.Contains(response, "fail") {
		message := utils.GetFieldStrFromJson(response, "message")
		status := utils.GetFieldStrFromJson(response, "status")
		logs.GetLogger().Fatal(status, ": ", message)
	}

	jwtToken := utils.GetFieldMapFromJson(response, "data")
	if jwtToken == nil {
		logs.GetLogger().Fatal("Error: fail to connect swan api")
	}

	jwt := jwtToken["jwt"].(string)

	swanClient := &SwanClient{
		ApiUrl: mainConf.SwanApiUrl,
		ApiKey: mainConf.SwanApiKey,
		Token:  jwt,
	}

	return swanClient
}

func (swanClient *SwanClient) SwanGetOfflineDeals(minerFid, status string, limit ...string) []model.OfflineDeal {
	rowLimit := strconv.Itoa(GET_OFFLINEDEAL_LIMIT_DEFAULT)
	if len(limit) > 0 {
		rowLimit = limit[0]
	}

	urlStr := swanClient.ApiUrl + "/offline_deals/" + minerFid + "?deal_status=" + status + "&limit=" + rowLimit + "&offset=0"
	response := HttpGet(urlStr, swanClient.Token, "")
	getOfflineDealResponse := GetOfflineDealResponse{}
	err := json.Unmarshal([]byte(response), &getOfflineDealResponse)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	if strings.ToUpper(getOfflineDealResponse.Status) != RESPONSE_STATUS_SUCCESS {
		logs.GetLogger().Error("Get offline deal with status ", status, " failed")
		return nil
	}

	return getOfflineDealResponse.Data.Deal
}

func (swanClient *SwanClient) SwanUpdateOfflineDealStatus(dealId int, status string, statusInfo ...string) bool {
	if len(status) == 0 {
		logs.GetLogger().Error("Please provide status")
		return false
	}

	apiUrl := swanClient.ApiUrl + "/my_miner/deals/" + strconv.Itoa(dealId)

	params := url.Values{}
	params.Add("status", status)

	if len(statusInfo) > 0 {
		params.Add("note", statusInfo[0])
	}

	if len(statusInfo) > 1 {
		params.Add("file_path", statusInfo[1])
	}

	if len(statusInfo) > 2 {
		params.Add("file_size", statusInfo[2])
	}

	response := HttpPut(apiUrl, swanClient.Token, strings.NewReader(params.Encode()))

	updateOfflineDealResponse := &UpdateOfflineDealResponse{}
	err := json.Unmarshal([]byte(response), updateOfflineDealResponse)
	if err != nil {
		logs.GetLogger().Error(err)
		return false
	}

	if strings.ToUpper(updateOfflineDealResponse.Status) != RESPONSE_STATUS_SUCCESS {
		logs.GetLogger().Error("Update offline deal with status ", status, " failed.", updateOfflineDealResponse.Data.Message)
		return false
	}

	return true
}

type SwanCreateTaskResponse struct {
	Data    SwanCreateTaskResponseData `json:"data"`
	Status  string                     `json:"status"`
	Message string                     `json:"message"`
}

type SwanCreateTaskResponseData struct {
	Filename string `json:"filename"`
	Uuid     string `json:"uuid"`
}

func (swanClient *SwanClient) SwanCreateTask(task model.Task, csvFilePath string) (*SwanCreateTaskResponse, error) {
	apiUrl := swanClient.ApiUrl + "/tasks"

	params := map[string]string{}
	params["task_name"] = task.TaskName
	params["curated_dataset"] = task.CuratedDataset
	params["description"] = task.Description
	params["is_public"] = strconv.Itoa(task.IsPublic)

	params["type"] = *task.Type

	if task.MinerFid != nil {
		params["miner_id"] = *task.MinerFid
	}
	params["fast_retrieval"] = strconv.FormatBool(task.FastRetrievalBool)
	params["bid_mode"] = strconv.Itoa(*task.BidMode)
	params["max_price"] = (*task.MaxPrice).String()
	params["expire_days"] = strconv.Itoa(*task.ExpireDays)

	response, err := HttpPostFile(apiUrl, swanClient.Token, params, "file", csvFilePath)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	swanCreateTaskResponse := &SwanCreateTaskResponse{}
	err = json.Unmarshal([]byte(response), swanCreateTaskResponse)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if swanCreateTaskResponse.Status != constants.SWAN_API_STATUS_SUCCESS {
		err := fmt.Errorf("error:%s,%s", swanCreateTaskResponse.Status, swanCreateTaskResponse.Message)
		logs.GetLogger().Error(err)
		return nil, err
	}

	return swanCreateTaskResponse, nil
}

type GetTaskResult struct {
	Data   GetTaskResultData `json:"data"`
	Status string            `json:"status"`
}

type GetTaskResultData struct {
	Task           []model.Task `json:"task"`
	TotalItems     int          `json:"total_items"`
	TotalTaskCount int          `json:"total_task_count"`
}

func (swanClient *SwanClient) GetTasks() ([]model.Task, error) {
	apiUrl := swanClient.ApiUrl + "/tasks"
	logs.GetLogger().Info("Getting My swan tasks info")
	response := HttpGet(apiUrl, swanClient.Token, "")

	if response == "" {
		err := errors.New("failed to get tasks from swan")
		logs.GetLogger().Error(err)
		return nil, err
	}

	//logs.GetLogger().Info(response)

	getTaskResult := &GetTaskResult{}
	err := json.Unmarshal([]byte(response), getTaskResult)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if getTaskResult.Status != constants.SWAN_API_STATUS_SUCCESS {
		err := fmt.Errorf("error:%s", getTaskResult.Status)
		logs.GetLogger().Error(err)
		return nil, err
	}

	return getTaskResult.Data.Task, nil
}

func (swanClient *SwanClient) GetAssignedTasks() ([]model.Task, error) {
	tasks, err := swanClient.GetTasks()
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if len(tasks) == 0 {
		return nil, nil
	}

	result := []model.Task{}

	for _, task := range tasks {
		if task.Status == constants.TASK_STATUS_ASSIGNED && task.MinerFid != nil {
			result = append(result, task)
		}
	}

	return result, nil
}

type GetOfflineDealsByTaskUuidResult struct {
	Data   GetOfflineDealsByTaskUuidResultData `json:"data"`
	Status string                              `json:"status"`
}
type GetOfflineDealsByTaskUuidResultData struct {
	AverageBid       string              `json:"average_bid"`
	BidCount         int                 `json:"bid_count"`
	DealCompleteRate string              `json:"deal_complete_rate"`
	Deal             []model.OfflineDeal `json:"deal"`
	Miner            model.Miner         `json:"miner"`
	Task             model.Task          `json:"task"`
}

func (swanClient *SwanClient) GetOfflineDealsByTaskUuid(taskUuid string) (*GetOfflineDealsByTaskUuidResult, error) {
	if len(taskUuid) == 0 {
		err := fmt.Errorf("please provide task uuid")
		logs.GetLogger().Error(err)
		return nil, err
	}
	apiUrl := swanClient.ApiUrl + "/tasks/" + taskUuid
	logs.GetLogger().Info("Getting My swan tasks info")
	response := HttpGet(apiUrl, swanClient.Token, "")

	if response == "" {
		err := errors.New("failed to get tasks from swan")
		logs.GetLogger().Error(err)
		return nil, err
	}
	//logs.GetLogger().Info(response)

	getOfflineDealsByTaskUuidResult := &GetOfflineDealsByTaskUuidResult{}
	err := json.Unmarshal([]byte(response), getOfflineDealsByTaskUuidResult)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if getOfflineDealsByTaskUuidResult.Status != constants.SWAN_API_STATUS_SUCCESS {
		err := fmt.Errorf("error:%s", getOfflineDealsByTaskUuidResult.Status)
		logs.GetLogger().Error(err)
		return nil, err
	}

	return getOfflineDealsByTaskUuidResult, nil
}

func (swanClient *SwanClient) SwanUpdateTaskByUuid(taskUuid string, minerFid string, csvFilePath string) string {
	apiUrl := swanClient.ApiUrl + "/uuid_tasks/" + taskUuid
	params := map[string]string{}
	params["miner_fid"] = minerFid

	response, err := HttpPutFile(apiUrl, swanClient.Token, params, "file", csvFilePath)
	if err != nil {
		logs.GetLogger().Error(err)
		return ""
	}

	return response
}

func (swanClient *SwanClient) UpdateAssignedTask(taskUuid, status, csvFilePath string) (*SwanCreateTaskResponse, error) {
	apiUrl := swanClient.ApiUrl + "/tasks/" + taskUuid
	logs.GetLogger().Info("Updating Swan task")
	params := map[string]string{}
	params["status"] = status

	response, err := HttpPutFile(apiUrl, swanClient.Token, params, "file", csvFilePath)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	swanCreateTaskResponse := &SwanCreateTaskResponse{}
	err = json.Unmarshal([]byte(response), swanCreateTaskResponse)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if swanCreateTaskResponse.Status != constants.SWAN_API_STATUS_SUCCESS {
		err := fmt.Errorf("error:%s,%s", swanCreateTaskResponse.Status, swanCreateTaskResponse.Message)
		logs.GetLogger().Error(err)
		return nil, err
	}

	return swanCreateTaskResponse, nil
}
