package command

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"math"
	"math/big"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/filswan/go-swan-client/config"
	"github.com/google/uuid"

	"github.com/filswan/go-swan-lib/logs"
	"github.com/filswan/go-swan-lib/utils"

	"github.com/filswan/go-swan-lib/client/lotus"
	"github.com/filswan/go-swan-lib/client/swan"
	libconstants "github.com/filswan/go-swan-lib/constants"
	libmodel "github.com/filswan/go-swan-lib/model"
)

type CmdAutoBidDeal struct {
	SwanApiUrl             string //required
	SwanApiKey             string //required when SwanJwtToken is not provided
	SwanAccessToken        string //required when SwanJwtToken is not provided
	SwanToken              string //required when SwanApiKey and SwanAccessToken are not provided
	LotusClientApiUrl      string //required
	LotusClientAccessToken string //required
	SenderWallet           string //required
	OutputDir              string //required
	DealSourceIds          []int  //required
}

func GetCmdAutoDeal(outputDir *string) *CmdAutoBidDeal {
	cmdAutoBidDeal := &CmdAutoBidDeal{
		SwanApiUrl:             config.GetConfig().Main.SwanApiUrl,
		SwanApiKey:             config.GetConfig().Main.SwanApiKey,
		SwanAccessToken:        config.GetConfig().Main.SwanAccessToken,
		LotusClientApiUrl:      config.GetConfig().Lotus.ClientApiUrl,
		LotusClientAccessToken: config.GetConfig().Lotus.ClientAccessToken,
		SenderWallet:           config.GetConfig().Sender.Wallet,
	}

	cmdAutoBidDeal.DealSourceIds = append(cmdAutoBidDeal.DealSourceIds, libconstants.TASK_SOURCE_ID_SWAN)
	cmdAutoBidDeal.DealSourceIds = append(cmdAutoBidDeal.DealSourceIds, libconstants.TASK_SOURCE_ID_SWAN_CLIENT)

	if !utils.IsStrEmpty(outputDir) {
		cmdAutoBidDeal.OutputDir = *outputDir
	} else {
		cmdAutoBidDeal.OutputDir = filepath.Join(config.GetConfig().Sender.OutputDir, time.Now().Format("2006-01-02_15:04:05")) + "_" + uuid.NewString()
	}

	return cmdAutoBidDeal
}

func SendAutoBidDealsLoopByConfig(outputDir string) error {
	cmdAutoBidDeal := GetCmdAutoDeal(&outputDir)
	err := cmdAutoBidDeal.SendAutoBidDealsLoop()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}

func (cmdAutoBidDeal *CmdAutoBidDeal) SendAutoBidDealsLoop() error {
	err := utils.CreateDirIfNotExists(cmdAutoBidDeal.OutputDir, DIR_NAME_OUTPUT)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	for {
		err := cmdAutoBidDeal.SendAutoBidDeals()
		if err != nil {
			logs.GetLogger().Error(err)
			continue
		}

		logs.GetLogger().Info("sleeping...")
		time.Sleep(time.Second * 30)
	}
}

func (cmdAutoBidDeal *CmdAutoBidDeal) SendAutoBidDeals() error {
	for _, sourceId := range cmdAutoBidDeal.DealSourceIds {
		logs.GetLogger().Info("send auto bid deals for souce:", sourceId)
		_, _, err := cmdAutoBidDeal.sendAutoBidDealsBySourceId(sourceId)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
	}

	return nil
}

func (cmdAutoBidDeal *CmdAutoBidDeal) sendAutoBidDealsBySourceId(sourceId int) ([]*string, [][]*libmodel.FileDesc, error) {
	swanClient, err := swan.GetClient(cmdAutoBidDeal.SwanApiUrl, cmdAutoBidDeal.SwanApiKey, cmdAutoBidDeal.SwanAccessToken, cmdAutoBidDeal.SwanToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	params := swan.GetOfflineDealsByStatusParams{
		DealStatus: libconstants.OFFLINE_DEAL_STATUS_ASSIGNED,
		ForMiner:   false,
		SourceId:   &sourceId,
	}
	assignedOfflineDeals, err := swanClient.GetOfflineDealsByStatus(params)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	var jsonFilepaths []*string
	var tasksFileDescs [][]*libmodel.FileDesc

	taskUuids := []string{}
	for _, offlineDeal := range assignedOfflineDeals {
		uuidExists := false
		for _, taskUuid := range taskUuids {
			if *offlineDeal.TaskUuid == taskUuid {
				uuidExists = true
				break
			}
		}
		if !uuidExists {
			taskUuids = append(taskUuids, *offlineDeal.TaskUuid)
		}
	}

	for _, taskUuid := range taskUuids {
		offlineDeals := []*libmodel.OfflineDeal{}
		for _, offlineDeal := range assignedOfflineDeals {
			if *offlineDeal.TaskUuid == taskUuid {
				offlineDeals = append(offlineDeals, offlineDeal)
			}
		}

		jsonFilepath, fileDescs, err := cmdAutoBidDeal.sendAutoBidDeals4Task(offlineDeals)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, nil, err
		}

		jsonFilepaths = append(jsonFilepaths, jsonFilepath)
		tasksFileDescs = append(tasksFileDescs, fileDescs)
	}

	return jsonFilepaths, tasksFileDescs, nil
}

func (cmdAutoBidDeal *CmdAutoBidDeal) SendAutoBidDealsByTaskUuid(taskUuid string) (*string, []*libmodel.FileDesc, error) {
	swanClient, err := swan.GetClient(cmdAutoBidDeal.SwanApiUrl, cmdAutoBidDeal.SwanApiKey, cmdAutoBidDeal.SwanAccessToken, cmdAutoBidDeal.SwanToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	params := swan.GetOfflineDealsByStatusParams{
		DealStatus: libconstants.OFFLINE_DEAL_STATUS_ASSIGNED,
		ForMiner:   false,
		TaskUuid:   &taskUuid,
	}
	assignedOfflineDeals, err := swanClient.GetOfflineDealsByStatus(params)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	jsonFilepath, fileDescs, err := cmdAutoBidDeal.sendAutoBidDeals4Task(assignedOfflineDeals)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	return jsonFilepath, fileDescs, nil
}

func (cmdAutoBidDeal *CmdAutoBidDeal) sendAutoBidDeals4Task(assignedOfflineDeals []*libmodel.OfflineDeal) (*string, []*libmodel.FileDesc, error) {
	if len(assignedOfflineDeals) == 0 {
		logs.GetLogger().Info("no offline deals to be sent")
		return nil, nil, nil
	}

	swanClient, err := swan.GetClient(cmdAutoBidDeal.SwanApiUrl, cmdAutoBidDeal.SwanApiKey, cmdAutoBidDeal.SwanAccessToken, cmdAutoBidDeal.SwanToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	fileDescs := []*libmodel.FileDesc{}

	for _, assignedOfflineDeal := range assignedOfflineDeals {
		fileDesc, err := cmdAutoBidDeal.sendAutobidDeal(assignedOfflineDeal)
		if err != nil {
			logs.GetLogger().Error(err)
			continue
		}

		if fileDesc == nil {
			continue
		}

		fileDescs = append(fileDescs, fileDesc)

		var cost string
		dealInfo, err := cmdAutoBidDeal.CheckDealStatus(fileDesc.Deals[0].DealCid)
		if err != nil {
			logs.GetLogger().Error(err)
			cost = "fail"
		} else {
			cost = dealInfo.CostComputed
		}

		updateOfflineDealParams := swan.UpdateOfflineDealParams{
			DealId:     assignedOfflineDeal.Id,
			DealCid:    &fileDesc.Deals[0].DealCid,
			Status:     libconstants.OFFLINE_DEAL_STATUS_CREATED,
			StartEpoch: &fileDesc.Deals[0].StartEpoch,
			Cost:       &cost,
		}

		err = swanClient.UpdateOfflineDeal(updateOfflineDealParams)
		if err != nil {
			logs.GetLogger().Error(err)
			continue
		}
	}

	jsonFileName := *assignedOfflineDeals[0].TaskName + JSON_FILE_NAME_DEAL_AUTO
	csvFileName := *assignedOfflineDeals[0].TaskName + CSV_FILE_NAME_DEAL_AUTO
	filepath, err := WriteCarFilesToFiles(fileDescs, cmdAutoBidDeal.OutputDir, jsonFileName, csvFileName)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	return filepath, fileDescs, nil
}

func (cmdAutoBidDeal *CmdAutoBidDeal) sendAutobidDeal(offlineDeal *libmodel.OfflineDeal) (*libmodel.FileDesc, error) {
	offlineDeal.DealCid = strings.Trim(offlineDeal.DealCid, " ")
	if len(offlineDeal.DealCid) != 0 {
		logs.GetLogger().Info("deal already be sent, task:%s, deal:%d", *offlineDeal.TaskUuid, offlineDeal.Id)
		return nil, nil
	}

	if offlineDeal.TaskType == nil {
		err := fmt.Errorf("task type is nil")
		logs.GetLogger().Error(err)
		return nil, err
	}

	if offlineDeal.FastRetrieval == nil {
		err := fmt.Errorf("FastRetrieval is nil")
		logs.GetLogger().Error(err)
		return nil, err
	}

	if offlineDeal.MaxPrice == nil {
		err := fmt.Errorf("MaxPrice is nil")
		logs.GetLogger().Error(err)
		return nil, err
	}

	if offlineDeal.Duration == nil {
		err := fmt.Errorf("duration is nil")
		logs.GetLogger().Error(err)
		return nil, err
	}

	dealConfig := libmodel.DealConfig{
		VerifiedDeal:     *offlineDeal.TaskType == libconstants.TASK_TYPE_VERIFIED,
		FastRetrieval:    *offlineDeal.FastRetrieval == libconstants.TASK_FAST_RETRIEVAL_YES,
		SkipConfirmation: true,
		MaxPrice:         *offlineDeal.MaxPrice,
		StartEpoch:       int64(offlineDeal.StartEpoch),
		MinerFid:         offlineDeal.MinerFid,
		SenderWallet:     cmdAutoBidDeal.SenderWallet,
		Duration:         int(*offlineDeal.Duration),
		TransferType:     libconstants.LOTUS_TRANSFER_TYPE_MANUAL,
		PayloadCid:       offlineDeal.PayloadCid,
		PieceCid:         offlineDeal.PieceCid,
		FileSize:         offlineDeal.CarFileSize,
	}

	msg := fmt.Sprintf("send deal for task:%d,%s, deal:%d", offlineDeal.TaskId, *offlineDeal.TaskUuid, offlineDeal.Id)
	logs.GetLogger().Info(msg)

	lotusClient, err := lotus.LotusGetClient(cmdAutoBidDeal.LotusClientApiUrl, cmdAutoBidDeal.LotusClientAccessToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if offlineDeal.TaskUuid == nil {
		err := fmt.Errorf("task uuid is nil")
		logs.GetLogger().Error(err)
		return nil, err
	}

	for i := 0; i < 60; i++ {
		//change start epoch to ensure that the deal cid is different
		dealConfig.StartEpoch = dealConfig.StartEpoch - (int64)(i)

		dealCid, err := lotusClient.LotusClientStartDeal(&dealConfig)
		if err != nil {
			logs.GetLogger().Error("tried ", i+1, " times,", err)

			if strings.Contains(err.Error(), "already tracking identifier") {
				continue
			} else {
				break
			}
		}
		if dealCid == nil {
			logs.GetLogger().Info("no deal CID returned")
			continue
		}

		dealInfo := &libmodel.DealInfo{
			MinerFid:   offlineDeal.MinerFid,
			DealCid:    *dealCid,
			StartEpoch: int(dealConfig.StartEpoch),
		}

		fileDesc := libmodel.FileDesc{
			Uuid:        *offlineDeal.TaskUuid,
			CarFileMd5:  offlineDeal.Md5Local,
			CarFileUrl:  offlineDeal.CarFileUrl,
			CarFileSize: offlineDeal.CarFileSize,
			PayloadCid:  offlineDeal.PayloadCid,
			PieceCid:    offlineDeal.PieceCid,
			SourceId:    offlineDeal.SourceId,
			Deals:       []*libmodel.DealInfo{},
		}
		fileDesc.Deals = append(fileDesc.Deals, dealInfo)

		logs.GetLogger().Info("deal sent successfully, task:", offlineDeal.TaskId, ", uuid:", *offlineDeal.TaskUuid, ", deal:", offlineDeal.Id, ", task name:", offlineDeal.TaskName, ", deal CID:", dealInfo.DealCid, ", start epoch:", dealInfo.StartEpoch, ", miner:", dealInfo.MinerFid)
		return &fileDesc, nil
	}

	err = fmt.Errorf("failed to send deal for task:%d,uuid:%s,deal:%d", offlineDeal.TaskId, *offlineDeal.TaskUuid, offlineDeal.Id)
	logs.GetLogger().Error(err)
	return nil, err
}

func (cmdAutoBidDeal *CmdAutoBidDeal) CheckDealStatus(dealCid string) (*lotus.ClientDealCostStatus, error) {
	lotusClient, err := lotus.LotusGetClient(cmdAutoBidDeal.LotusClientApiUrl, cmdAutoBidDeal.LotusClientAccessToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	dealCostInfo, err := lotusClient.LotusClientGetDealInfo(dealCid)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	return dealCostInfo, err
}

func SendRpcReqAndResp(chainId, params string) (result []byte, err error) {
	urls, ok := publicChain[chainId]
	if !ok {
		return nil, errors.New(fmt.Sprintf("not support chainId: %s", chainId))
	}

	for _, u := range urls {
		copyUrl := u
		result, err = doReq(copyUrl, params)
		if err != nil {
			if len(urls) > 0 {
				fmt.Printf("occur error, msg: %v, retry it ...", err)
				continue
			}
			fmt.Printf("occur error, msg: %v", err)
		}
		break
	}
	return
}

func doReq(reqUrl, params string) ([]byte, error) {
	client := http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Post(reqUrl, "application/json", strings.NewReader(params))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func QueryChainInfo(chain string, height int64, address string) (ChainInfo, error) {
	if utils.IsStrEmpty(&address) {
		return QueryHeight(chain)
	} else {
		return QueryBalance(chain, height, address)
	}
}

func QueryHeight(chain string) (info ChainInfo, err error) {
	urls, ok := chainUrlMap[chain]
	if !ok {
		return info, errors.New(fmt.Sprintf("not support chainId: %s", chain))
	}
	var rpcParam rpcReq
	switch chain {
	case "ETH", "BNB", "MATIC", "FTM", "xDAI", "IOTX", "BOBA", "EVMOS", "AVAX", "FUSE", "JEWEL", "TUS":
		rpcParam.Jsonrpc = "2.0"
		rpcParam.Method = "eth_blockNumber"
		rpcParam.Params = []interface{}{}
		rpcParam.Id = 83
	case "ONE":
		rpcParam.Jsonrpc = "2.0"
		rpcParam.Method = "hmyv2_blockNumber"
		rpcParam.Id = 1
		rpcParam.Params = []interface{}{}
	}

	var data []byte
	for _, u := range urls {
		copyUrl := u
		rpcParamBytes, err := json.Marshal(rpcParam)
		if err != nil {
			logs.GetLogger().Errorf("generate req params bytes failed,error: %v", err)
			continue
		}
		data, err = doReq(copyUrl, string(rpcParamBytes))
		if err != nil {
			logs.GetLogger().Errorf("request url: %s failed, error: %v", copyUrl, err)
			if len(urls) > 0 {
				logs.GetLogger().Warnf("retry it by other request")
				continue
			}
			break
		}
		break
	}

	height := utils.GetFieldFromJson(data, "result")
	errorInfo := GetFieldMapFromJsonByError(data)

	if errorInfo != nil {
		errCode := int(errorInfo["code"].(float64))
		errMsg := errorInfo["message"].(string)
		logs.GetLogger().Errorf("get %s height failed, code: %d error: %s", chain, errCode, errMsg)
		return
	}

	switch height.(type) {
	case float64:
		info.Height = uint64(height.(float64))
	case string:
		hex := height.(string)
		val := hex[2:]
		var data uint64
		data, err = strconv.ParseUint(val, 16, 64)
		if err != nil {
			logs.GetLogger().Errorf("convert height value failed, error: %v", err)
			return
		}
		info.Height = data
	}

	return
}

func QueryBalance(chain string, height int64, address string) (info ChainInfo, err error) {
	urls, ok := chainUrlMap[chain]
	if !ok {
		return info, errors.New(fmt.Sprintf("not support chainId: %s", chain))
	}

	// IOTX need change wallet to eth wallet
	var rpcParam rpcReq
	switch chain {
	case "ETH", "BNB", "MATIC", "FTM", "xDAI", "IOTX", "BOBA", "EVMOS", "AVAX", "FUSE", "JEWEL", "TUS":
		rpcParam.Jsonrpc = "2.0"
		rpcParam.Method = "eth_getBalance"
		chainHeight := "latest"
		rpcParam.Id = 1
		if height != 0 {
			chainHeight = strconv.Itoa(int(height))
			info.Height = uint64(height)
		} else {
			queryHeight, err := QueryHeight(chain)
			if err != nil {
				return queryHeight, err
			}
			info.Height = queryHeight.Height
		}
		rpcParam.Params = []interface{}{address, chainHeight}

	case "ONE":
		rpcParam.Jsonrpc = "2.0"
		rpcParam.Method = "hmyv2_getBalanceByBlockNumber"
		rpcParam.Id = 1
		if height != 0 {
			info.Height = uint64(height)
		} else {
			queryHeight, err := QueryHeight(chain)
			if err != nil {
				return queryHeight, err
			}
			info.Height = queryHeight.Height
		}
		rpcParam.Params = []interface{}{address, info.Height}
	}

	info.Address = address
	var data []byte
	for _, u := range urls {
		copyUrl := u
		rpcParamBytes, err := json.Marshal(rpcParam)
		if err != nil {
			logs.GetLogger().Errorf("generate req params bytes failed,error: %v", err)
			continue
		}
		data, err = doReq(copyUrl, string(rpcParamBytes))
		if err != nil {
			logs.GetLogger().Errorf("request url: %s failed, error: %v", copyUrl, err)
			if len(urls) > 0 {
				logs.GetLogger().Warnf("retry it by other request")
				continue
			}
			break
		}
	}

	balance := utils.GetFieldFromJson(data, "result")
	errorInfo := GetFieldMapFromJsonByError(data)

	if errorInfo != nil {
		errCode := int(errorInfo["code"].(float64))
		errMsg := errorInfo["message"].(string)
		logs.GetLogger().Errorf("get %s balance failed, code: %d error: %s", chain, errCode, errMsg)
		return
	}

	switch balance.(type) {
	case float64:
		fbalance := new(big.Float)
		fbalance.SetFloat64(balance.(float64))
		info.Balance = new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18)))
	case string:
		hex := balance.(string)
		val := hex[2:]
		var data uint64
		data, err = strconv.ParseUint(val, 16, 64)
		if err != nil {
			logs.GetLogger().Errorf("convert balance value failed, error: %v", err)
			return
		}
		fbalance := new(big.Float)
		fbalance.SetUint64(data)
		info.Balance = new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18)))
	}
	return
}

type ChainInfo struct {
	Height  uint64
	Balance *big.Float
	Address string
}

// {"jsonrpc":"2.0","method":"eth_getBalance","params":["0x407d73d8a49eeb85d32cf465507dd71d507100c1", "latest"],"id":1}
type rpcReq struct {
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      int           `json:"id"`
}

type rpcResp struct {
	Id      int    `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Result  string `json:"result"`
	Error   struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func GetFieldMapFromJsonByError(jsonBytes []byte) map[string]interface{} {
	fieldVal := utils.GetFieldFromJson(jsonBytes, "error")
	if fieldVal == nil {
		return nil
	}

	switch fieldValType := fieldVal.(type) {
	case map[string]interface{}:
		return fieldValType
	default:
		return nil
	}
}
