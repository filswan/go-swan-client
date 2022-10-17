package command

import (
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"path/filepath"
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
		cmdAutoBidDeal.OutputDir = filepath.Join(*outputDir, time.Now().Format("2006-01-02_15:04:05")) + "_" + uuid.NewString()
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

		_, jsonFilepath, fileDescs, err := cmdAutoBidDeal.sendAutoBidDeals4Task(offlineDeals, nil)
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

	_, jsonFilepath, fileDescs, err := cmdAutoBidDeal.sendAutoBidDeals4Task(assignedOfflineDeals, nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	return jsonFilepath, fileDescs, nil
}

func (cmdAutoBidDeal *CmdAutoBidDeal) sendAutoBidDeals4Task(assignedOfflineDeals []*libmodel.OfflineDeal, carSourceInfoMap map[string]*libmodel.FileDesc) (int, *string, []*libmodel.FileDesc, error) {
	if len(assignedOfflineDeals) == 0 {
		logs.GetLogger().Info("no offline deals to be sent")
		return 0, nil, nil, nil
	}

	swanClient, err := swan.GetClient(cmdAutoBidDeal.SwanApiUrl, cmdAutoBidDeal.SwanApiKey, cmdAutoBidDeal.SwanAccessToken, cmdAutoBidDeal.SwanToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return 0, nil, nil, err
	}

	var successCount int
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

		if carSourceInfoMap != nil {
			if carSourceInfo, ok := carSourceInfoMap[fileDesc.PayloadCid]; ok {
				fileDesc.SourceFileName = carSourceInfo.SourceFileName
				fileDesc.SourceFilePath = carSourceInfo.SourceFilePath
				fileDesc.SourceFileMd5 = carSourceInfo.SourceFileMd5
				fileDesc.SourceFileSize = carSourceInfo.SourceFileSize
				fileDesc.CarFileName = carSourceInfo.CarFileName
				fileDesc.CarFilePath = carSourceInfo.CarFilePath
				fileDesc.CarFileMd5 = carSourceInfo.CarFileMd5
			}
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
		successCount += 1
	}

	jsonFileName := *assignedOfflineDeals[0].TaskName + JSON_FILE_NAME_DEAL_AUTO
	csvFileName := *assignedOfflineDeals[0].TaskName + CSV_FILE_NAME_DEAL_AUTO
	filePath, err := WriteCarFilesToFiles(fileDescs, cmdAutoBidDeal.OutputDir, jsonFileName, csvFileName)
	if err != nil {
		logs.GetLogger().Error(err)
		return successCount, nil, nil, err
	}

	return successCount, filePath, fileDescs, nil
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

	msg := fmt.Sprintf("send deal for taskName: %s, PayloadCid:%s", *offlineDeal.TaskName, offlineDeal.PayloadCid)
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

		logs.GetLogger().Infof("deal sent successfully, taskName: %s,dealId: %d,dealCid: %s,minerId: %s,startEpoch: %d .", *offlineDeal.TaskName, offlineDeal.Id, dealInfo.DealCid, dealInfo.MinerFid, dealInfo.StartEpoch)
		return &fileDesc, nil
	}

	err = fmt.Errorf("failed to send deal for taskName: %d,dealCid: %s", offlineDeal.TaskName, offlineDeal.DealCid)
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

func (cmdAutoBidDeal *CmdAutoBidDeal) SendAutoBidDealsBySwanClientSourceId(inputDir string, taskUuid string, total int) {
	swanClient, err := swan.GetClient(cmdAutoBidDeal.SwanApiUrl, cmdAutoBidDeal.SwanApiKey, cmdAutoBidDeal.SwanAccessToken, cmdAutoBidDeal.SwanToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}

	var fileDescs []*libmodel.FileDesc
	if strings.HasSuffix(inputDir, "json") {
		fileDescs, err = ReadFileDescsFromJsonFile(inputDir, "")
		if err != nil {
			logs.GetLogger().Error(err)
			return
		}
	}
	if strings.HasSuffix(inputDir, "csv") {
		fileDescs, err = ReadFileFromCsvFile(inputDir, "")
		if err != nil {
			logs.GetLogger().Error(err)
			return
		}
	}

	var carSourceInfoMap = make(map[string]*libmodel.FileDesc)
	for _, desc := range fileDescs {
		carSourceInfoMap[desc.PayloadCid] = desc
	}

	sourceId := libconstants.TASK_SOURCE_ID_SWAN_CLIENT
	params := swan.GetOfflineDealsByStatusParams{
		DealStatus: libconstants.OFFLINE_DEAL_STATUS_ASSIGNED,
		ForMiner:   false,
		SourceId:   &sourceId,
	}

	var totalCount, successCount int
	tick := time.Tick(10 * time.Second)
	for {
		select {
		case <-tick:
			assignedOfflineDeals, err := swanClient.GetOfflineDealsByStatus(params)
			if err != nil {
				logs.GetLogger().Error(err)
				continue
			}
			var offlineDeals []*libmodel.OfflineDeal
			for _, offlineDeal := range assignedOfflineDeals {
				if *offlineDeal.TaskUuid == taskUuid {
					offlineDeals = append(offlineDeals, offlineDeal)
				}
			}
			totalCount += len(offlineDeals)
			success, _, _, err := cmdAutoBidDeal.sendAutoBidDeals4Task(offlineDeals, carSourceInfoMap)
			if err != nil {
				logs.GetLogger().Error(err)
				continue
			}
			successCount += success
			if totalCount == total {
				goto END
			}
		case <-time.After(30 * time.Minute):
			goto END
		}
	}
END:
	logs.GetLogger().Infof("auto send deal end,dealTotalCount: %d,successed: %d,failed: %d", total, successCount, total-successCount)
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
