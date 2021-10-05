package utils

import (
	"encoding/json"
	"go-swan-client/common/constants"
	"go-swan-client/config"
	"go-swan-client/logs"
	"go-swan-client/models"
	"net/url"
	"strconv"
	"strings"
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

type MinerResponse struct {
	Status  string       `json:"status"`
	Message string       `json:"message"`
	Data    models.Miner `json:"data"`
}

type GetOfflineDealResponse struct {
	Data   GetOfflineDealData `json:"data"`
	Status string             `json:"status"`
}

type GetOfflineDealData struct {
	Deal []models.OfflineDeal `json:"deal"`
}

type UpdateOfflineDealResponse struct {
	Data   UpdateOfflineDealData `json:"data"`
	Status string                `json:"status"`
}

type UpdateOfflineDealData struct {
	Deal    models.OfflineDeal `json:"deal"`
	Message string             `json:"message"`
}

func SwanGetClient() *SwanClient {
	mainConf := config.GetConfig().Main
	uri := mainConf.SwanApiUrl + "/user/api_keys/jwt"
	data := TokenAccessInfo{ApiKey: mainConf.SwanApiKey, AccessToken: mainConf.SwanAccessToken}
	response := HttpPostNoToken(uri, data)

	if strings.Contains(response, "fail") {
		message := GetFieldStrFromJson(response, "message")
		status := GetFieldStrFromJson(response, "status")
		logs.GetLogger().Fatal(status, ": ", message)
	}

	jwtToken := GetFieldMapFromJson(response, "data")
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

func (swanClient *SwanClient) SwanGetMiner(minerFid string) *MinerResponse {
	apiUrl := swanClient.ApiUrl + "/miner/info/" + minerFid

	response := HttpGetNoToken(apiUrl, "")
	minerResponse := &MinerResponse{}
	err := json.Unmarshal([]byte(response), minerResponse)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	return minerResponse
}

func (swanClient *SwanClient) SwanGetOfflineDeals(minerFid, status string, limit ...string) []models.OfflineDeal {
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

func (swanClient *SwanClient) SwanSendHeartbeatRequest(minerFid string) string {
	apiUrl := swanClient.ApiUrl + "/heartbeat"
	params := url.Values{}
	params.Add("miner_id", minerFid)

	response := HttpPost(apiUrl, swanClient.Token, strings.NewReader(params.Encode()))
	return response
}

func (swanClient *SwanClient) SwanUpdateTaskByUuid(taskUuid string, minerFid string) string {
	apiUrl := swanClient.ApiUrl + "/uuid_tasks/" + taskUuid
	params := url.Values{}
	params.Add("miner_fid", minerFid)

	response := HttpPut(apiUrl, swanClient.Token, params)

	return response
}

func (swanClient *SwanClient) SwanCreateTask(task models.Task, minerId *string, csvFilePath string) string {
	apiUrl := swanClient.ApiUrl + "/tasks"

	params := map[string]string{}
	params["task_name"] = task.TaskName
	params["curated_dataset"] = task.CuratedDataset
	params["description"] = task.Description
	params["is_public"] = strconv.FormatBool(task.IsPublic)
	if task.IsVerified {
		params["type"] = constants.TASK_TYPE_VERIFIED
	} else {
		params["type"] = constants.TASK_TYPE_REGULAR
	}

	if minerId != nil {
		params["miner_id"] = *minerId
	}

	response, err := HttpPostFile(apiUrl, swanClient.Token, params, csvFilePath)
	if err != nil {
		logs.GetLogger().Error(err)
		return ""
	}
	return response
}

func (swanClient *SwanClient) SwanUpdateMiner(miner models.Miner) string {
	apiUrl := swanClient.ApiUrl + "/miners/" + strconv.Itoa(miner.Id) + "/status"
	params := url.Values{}
	params.Add("price", strconv.FormatFloat(*miner.Price, 'E', -1, 64))
	params.Add("verified_price", strconv.FormatFloat(*miner.VerifiedPrice, 'E', -1, 64))
	params.Add("min_piece_size", *miner.MinPieceSize)
	params.Add("max_piece_size", *miner.MaxPieceSize)

	response := HttpPut(apiUrl, swanClient.Token, params)
	return response
}
