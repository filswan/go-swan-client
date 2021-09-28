package utils

import (
	"encoding/json"
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
	Deal []models.OfflineDeals `json:"deal""`
}

type UpdateOfflineDealResponse struct {
	Data   UpdateOfflineDealData `json:"data"`
	Status string                `json:"status"`
}

type UpdateOfflineDealData struct {
	Deal    models.OfflineDeals `json:"deal""`
	Message string              `json:"message"`
}

func GetSwanClient() *SwanClient {
	mainConf := config.GetConfig().Main
	uri := mainConf.SwanApiUrl + "/user/api_keys/jwt"
	data := TokenAccessInfo{ApiKey: mainConf.SwanApiKey, AccessToken: mainConf.SwanAccessToken}
	response := HttpPostNoToken(uri, data)

	if strings.Index(response, "fail") >= 0 {
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

func (self *SwanClient) GetMiner(minerFid string) *MinerResponse {
	apiUrl := self.ApiUrl + "/miner/info/" + minerFid

	response := HttpGetNoToken(apiUrl, "")
	minerResponse := &MinerResponse{}
	err := json.Unmarshal([]byte(response), minerResponse)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	return minerResponse
}

func (self *SwanClient) GetOfflineDeals(minerFid, status string, limit ...string) []models.OfflineDeals {
	rowLimit := strconv.Itoa(GET_OFFLINEDEAL_LIMIT_DEFAULT)
	if limit != nil && len(limit) > 0 {
		rowLimit = limit[0]
	}

	urlStr := self.ApiUrl + "/offline_deals/" + minerFid + "?deal_status=" + status + "&limit=" + rowLimit + "&offset=0"
	response := HttpGet(urlStr, self.Token, "")
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

func (self *SwanClient) UpdateOfflineDealStatus(dealId int, status string, statusInfo ...string) bool {
	if len(status) == 0 {
		logs.GetLogger().Error("Please provide status")
		return false
	}

	apiUrl := self.ApiUrl + "/my_miner/deals/" + strconv.Itoa(dealId)

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

	response := HttpPut(apiUrl, self.Token, strings.NewReader(params.Encode()))

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

func (self *SwanClient) SendHeartbeatRequest(minerFid string) string {
	apiUrl := self.ApiUrl + "/heartbeat"
	params := url.Values{}
	params.Add("miner_id", minerFid)

	response := HttpPost(apiUrl, self.Token, strings.NewReader(params.Encode()))
	return response
}

func (self *SwanClient) UpdateTaskByUuid(taskUuid string, minerFid string) string {
	apiUrl := self.ApiUrl + "/uuid_tasks/" + taskUuid
	params := url.Values{}
	params.Add("miner_fid", minerFid)

	response := HttpPut(apiUrl, self.Token, params)

	return response
}

func (self *SwanClient) PostTask(task models.Task) string {
	apiUrl := self.ApiUrl + "/tasks"

	params := url.Values{}
	params.Add("task_name", task.TaskName)
	params.Add("curated_dataset", task.CuratedDataset)
	params.Add("description", task.Description)
	params.Add("is_public", strconv.Itoa(task.IsPublic))
	params.Add("type", *task.Type)
	params.Add("miner_id", strconv.Itoa(*task.MinerId))

	response := HttpPost(apiUrl, self.Token, params)
	return response
}

func (self *SwanClient) UpdateMiner(miner models.Miner) string {
	apiUrl := self.ApiUrl + "/miners/" + strconv.Itoa(miner.Id) + "/status"
	params := url.Values{}
	params.Add("price", strconv.FormatFloat(*miner.Price, 'E', -1, 64))
	params.Add("verified_price", strconv.FormatFloat(*miner.VerifiedPrice, 'E', -1, 64))
	params.Add("min_piece_size", *miner.MinPieceSize)
	params.Add("max_piece_size", *miner.MaxPieceSize)

	response := HttpPut(apiUrl, self.Token, params)
	return response
}
