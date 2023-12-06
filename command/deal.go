package command

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/filswan/go-swan-client/config"
	"github.com/filswan/go-swan-lib/client/lotus"
	"github.com/filswan/go-swan-lib/client/swan"
	"github.com/filswan/go-swan-lib/client/web"
	libconstants "github.com/filswan/go-swan-lib/constants"
	"github.com/filswan/go-swan-lib/logs"
	libmodel "github.com/filswan/go-swan-lib/model"
	"github.com/filswan/go-swan-lib/utils"
	boost "github.com/filswan/swan-boost-lib/client"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CmdDeal struct {
	SwanApiUrl             string          //required
	SwanApiKey             string          //required when SwanJwtToken is not provided
	SwanAccessToken        string          //required when SwanJwtToken is not provided
	SwanToken              string          //required when SwanApiKey and SwanAccessToken are not provided
	LotusClientApiUrl      string          //required
	LotusClientAccessToken string          //required
	SenderWallet           string          //required
	MaxPrice               decimal.Decimal //required
	VerifiedDeal           bool            //required
	FastRetrieval          bool            //required
	SkipConfirmation       bool            //required
	Duration               int             //not necessary, when not provided use default value:1512000
	StartEpochHours        int             //required
	OutputDir              string          //required
	MinerFids              []string        //required
	MetadataJsonPath       string          //required
	MetadataCsvPath        string          //required
	StartDealTimeInterval  time.Duration   //required
	SwanRepo               string
	MarketVersion          string
}

func GetCmdDeal(outputDir *string, minerFids, metadataJsonPath, metadataCsvPath string) *CmdDeal {
	cmdDeal := &CmdDeal{
		SwanApiUrl:             config.GetConfig().Main.SwanApiUrl,
		SwanApiKey:             config.GetConfig().Main.SwanApiKey,
		SwanAccessToken:        config.GetConfig().Main.SwanAccessToken,
		LotusClientApiUrl:      config.GetConfig().Lotus.ClientApiUrl,
		LotusClientAccessToken: config.GetConfig().Lotus.ClientAccessToken,
		SenderWallet:           config.GetConfig().Sender.Wallet,
		VerifiedDeal:           config.GetConfig().Sender.VerifiedDeal,
		FastRetrieval:          config.GetConfig().Sender.FastRetrieval,
		SkipConfirmation:       config.GetConfig().Sender.SkipConfirmation,
		Duration:               config.GetConfig().Sender.Duration,
		StartEpochHours:        config.GetConfig().Sender.StartEpochHours,
		MinerFids:              []string{},
		MetadataJsonPath:       metadataJsonPath,
		MetadataCsvPath:        metadataCsvPath,
		StartDealTimeInterval:  config.GetConfig().Sender.StartDealTimeInterval,
		SwanRepo:               strings.TrimSpace(config.GetConfig().Main.SwanRepo),
		MarketVersion:          strings.TrimSpace(config.GetConfig().Main.MarketVersion),
	}

	minerFids = strings.Trim(minerFids, " ")
	if minerFids != "" {
		cmdDeal.MinerFids = strings.Split(minerFids, ",")
	}

	if !utils.IsStrEmpty(outputDir) {
		cmdDeal.OutputDir = *outputDir
	} else {
		cmdDeal.OutputDir = filepath.Join(*outputDir, time.Now().Format("2006-01-02_15:04:05")) + "_" + uuid.NewString()
	}

	maxPriceStr := strings.Trim(config.GetConfig().Sender.MaxPrice, " ")
	maxPrice, err := decimal.NewFromString(maxPriceStr)
	if err != nil {
		logs.GetLogger().Error("Failed to convert maxPrice(" + maxPriceStr + ") to decimal, MaxPrice:")
		return nil
	}
	cmdDeal.MaxPrice = maxPrice

	return cmdDeal
}

func SendDealsByConfig(outputDir, minerFid, metadataJsonPath, metadataCsvPath string) ([]*libmodel.FileDesc, error) {
	if metadataJsonPath == "" && metadataCsvPath == "" {
		err := fmt.Errorf("both metadataJsonPath and metadataCsvPath is nil")
		logs.GetLogger().Error(err)
		return nil, err
	}

	cmdDeal := GetCmdDeal(&outputDir, minerFid, metadataJsonPath, metadataCsvPath)
	fileDescs, err := cmdDeal.SendDeals()
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return fileDescs, nil
}

func (cmdDeal *CmdDeal) SendDeals() ([]*libmodel.FileDesc, error) {
	err := utils.CreateDirIfNotExists(cmdDeal.OutputDir, DIR_NAME_OUTPUT)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	fileDescs := make([]*libmodel.FileDesc, 0)
	var errMsg error

	if len(cmdDeal.MetadataJsonPath) > 0 {
		fileDescs, err = ReadFileDescsFromJsonFileByFullPath(cmdDeal.MetadataJsonPath)
		errMsg = fmt.Errorf("no car files read from:%s", cmdDeal.MetadataJsonPath)
	}
	if len(cmdDeal.MetadataCsvPath) > 0 {
		fileDescs, err = ReadFileFromCsvFileByFullPath(cmdDeal.MetadataCsvPath)
		errMsg = fmt.Errorf("no car files read from:%s", cmdDeal.MetadataJsonPath)
	}

	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	if len(fileDescs) == 0 {
		logs.GetLogger().Error(errMsg)
		return nil, err
	} else {
		errMsg = nil
	}

	swanClient, err := swan.GetClient(cmdDeal.SwanApiUrl, cmdDeal.SwanApiKey, cmdDeal.SwanAccessToken, cmdDeal.SwanToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	task, err := swanClient.GetTaskByUuid(fileDescs[0].Uuid)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if task.Data.Task.BidMode == nil || *task.Data.Task.BidMode != libconstants.TASK_BID_MODE_MANUAL {
		err := fmt.Errorf("auto_bid mode for task:%s is not manual, please check", task.Data.Task.TaskName)
		logs.GetLogger().Error(err)
		return nil, err
	}

	if cmdDeal.VerifiedDeal {
		isWalletVerified, err := cmdDeal.CheckDatacap(cmdDeal.SenderWallet)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}

		if !isWalletVerified {
			err := fmt.Errorf("task:%s is verified, but your wallet:%s is not verified", task.Data.Task.TaskName, cmdDeal.SenderWallet)
			logs.GetLogger().Error(err)
			return nil, err
		}
	}

	minerFids := make([]string, 0)
	for _, bid := range task.Data.Bids {
		if len(strings.TrimSpace(bid.MinerFid)) != 0 {
			minerFids = append(minerFids, bid.MinerFid)
		}
	}

	if len(cmdDeal.MinerFids) > 0 {
		if !minerFIdsIsExist(cmdDeal.MinerFids, minerFids) {
			return nil, fmt.Errorf("this task is not assigned to these miners: %+v, should be: %+v", cmdDeal.MinerFids, minerFids)
		}
	} else {
		cmdDeal.MinerFids = minerFids
	}
	logs.GetLogger().Infof("this task will send deal to these miners: %+v", cmdDeal.MinerFids)

	fileDescs, err = cmdDeal.sendDeals2Miner(task.Data.Task.TaskName, cmdDeal.OutputDir, fileDescs)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	_, err = swanClient.CreateOfflineDeals(fileDescs)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return fileDescs, nil
}

func (cmdDeal *CmdDeal) sendDeals2Miner(taskName string, outputDir string, fileDescs []*libmodel.FileDesc) ([]*libmodel.FileDesc, error) {
	lotusClient, err := lotus.LotusGetClient(cmdDeal.LotusClientApiUrl, cmdDeal.LotusClientAccessToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	dealSentNum := 0
	total := len(fileDescs) * len(cmdDeal.MinerFids)
	for _, fileDesc := range fileDescs {
		if fileDesc.CarFileSize <= 0 {
			logs.GetLogger().Error("File:" + fileDesc.CarFilePath + " %s is too small")
			continue
		}

		currentEpoch, err := lotusClient.LotusGetCurrentEpoch()
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}

		dealConfig := libmodel.DealConfig{
			VerifiedDeal:     cmdDeal.VerifiedDeal,
			FastRetrieval:    cmdDeal.FastRetrieval,
			SkipConfirmation: cmdDeal.SkipConfirmation,
			MaxPrice:         cmdDeal.MaxPrice,
			StartEpoch:       *currentEpoch + int64((cmdDeal.StartEpochHours+1)*libconstants.EPOCH_PER_HOUR),
			SenderWallet:     cmdDeal.SenderWallet,
			Duration:         int(cmdDeal.Duration),
			TransferType:     libconstants.LOTUS_TRANSFER_TYPE_MANUAL,
			PayloadCid:       fileDesc.PayloadCid,
			PieceCid:         fileDesc.PieceCid,
			FileSize:         fileDesc.CarFileSize,
			ClientRepo:       cmdDeal.SwanRepo,
		}

		logs.GetLogger().Info("File:", fileDesc.CarFilePath, ",current epoch:", *currentEpoch, ", start epoch:", dealConfig.StartEpoch)

		if len(cmdDeal.MinerFids) == 0 {
			cmdDeal.MinerFids = []string{}
			for _, deal := range fileDesc.Deals {
				cmdDeal.MinerFids = append(cmdDeal.MinerFids, deal.MinerFid)
			}
		}

		if len(cmdDeal.MinerFids) == 0 {
			err := fmt.Errorf("miner is required, you can set in command line or in metadata file")
			logs.GetLogger().Error(err)
			return nil, err
		}

		var deals []*libmodel.DealInfo
		for _, minerFid := range cmdDeal.MinerFids {
			dealConfig.MinerFid = minerFid

			var cost string
			var deal *libmodel.DealInfo
			if cmdDeal.MarketVersion == libconstants.MARKET_VERSION_2 {
				dealUuid, err := boost.GetClient(cmdDeal.SwanRepo).WithClient(lotusClient).StartDeal(&dealConfig)
				if err != nil {
					deals = append(deals, &libmodel.DealInfo{
						MinerFid:   dealConfig.MinerFid,
						DealCid:    "",
						StartEpoch: int(dealConfig.StartEpoch),
						Cost:       "fail",
					})
					logs.GetLogger().Infof("%d/%d deal sent failed, task name: %s, car file: %s, start epoch: %d, miner: %s, error: %v", len(deals), total, taskName, fileDesc.CarFilePath, dealConfig.StartEpoch, dealConfig.MinerFid, err)
					continue
				}
				deal = &libmodel.DealInfo{
					MinerFid:   dealConfig.MinerFid,
					DealCid:    dealUuid,
					StartEpoch: int(dealConfig.StartEpoch),
					Cost:       "0",
				}
			} else {
				dealCid, err := lotusStartDeal(lotusClient, &dealConfig)
				if err != nil {
					deals = append(deals, &libmodel.DealInfo{
						MinerFid:   dealConfig.MinerFid,
						DealCid:    "",
						StartEpoch: int(dealConfig.StartEpoch),
						Cost:       "fail",
					})
					logs.GetLogger().Infof("%d/%d deal sent failed, task name: %s, car file: %s, start epoch: %d, miner: %s, error: %v", len(deals), total, taskName, fileDesc.CarFilePath, dealConfig.StartEpoch, dealConfig.MinerFid, err)
					continue
				}
				if dealCid == nil {
					dealCid = new(string)
				} else {
					dealInfo, err := lotusClient.LotusClientGetDealInfo(*dealCid)
					if err != nil {
						logs.GetLogger().Error(err)
						cost = "fail"
					} else {
						cost = dealInfo.CostComputed
					}
				}
				deal = &libmodel.DealInfo{
					MinerFid:   dealConfig.MinerFid,
					DealCid:    *dealCid,
					StartEpoch: int(dealConfig.StartEpoch),
					Cost:       cost,
				}
			}

			deals = append(deals, deal)
			dealSentNum = dealSentNum + 1
			logs.GetLogger().Infof("%d/%d deal sent successfully, task name: %s, car file: %s, dealCID|dealUuid: %s, start epoch: %d, miner: %s", len(deals), total, taskName, fileDesc.CarFilePath, deal.DealCid, dealConfig.StartEpoch, dealConfig.MinerFid)
			if cmdDeal.StartDealTimeInterval > 0 {
				time.Sleep(cmdDeal.StartDealTimeInterval * time.Millisecond)
			}
		}

		fileDesc.Deals = deals
	}

	if cmdDeal.MarketVersion == libconstants.MARKET_VERSION_1 {
		fmt.Println(color.YellowString("You are using the MARKET(version=1.1 built-in Lotus) send deals, but it is deprecated, will remove soon. Please set [main.market_version=“1.2”]"))
	}

	logs.GetLogger().Infof("%d successful deal(s) and %d failed deal(s) has(ve) been sent for task: %s, minerID: %+v", dealSentNum, total-dealSentNum, taskName, cmdDeal.MinerFids)

	jsonFileName := taskName + JSON_FILE_NAME_DEAL
	csvFileName := taskName + CSV_FILE_NAME_DEAL
	_, err = WriteCarFilesToFiles(fileDescs, outputDir, jsonFileName, csvFileName)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return fileDescs, err
}

func lotusStartDeal(lotusClient *lotus.LotusClient, dealConfig *libmodel.DealConfig) (dealCid *string, err error) {
	pieceSize, epochPrice, err := boost.CheckDealConfig(lotusClient, dealConfig, true)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}
	return lotusClient.StartDeal(pieceSize, epochPrice, dealConfig)
}

func (cmdDeal *CmdDeal) CheckDatacap(address string) (bool, error) {

	var params []interface{}
	params = append(params, address)
	params = append(params, []interface{}{})

	jsonRpcParams := lotus.LotusJsonRpcParams{
		JsonRpc: lotus.LOTUS_JSON_RPC_VERSION,
		Method:  "Filecoin.StateVerifiedClientStatus",
		Params:  params,
		Id:      lotus.LOTUS_JSON_RPC_ID,
	}
	response, err := web.HttpPostNoToken(cmdDeal.LotusClientApiUrl, jsonRpcParams)
	if err != nil {
		logs.GetLogger().Error(err)
		return false, err
	}

	result := utils.GetFieldStrFromJson(response, "result")
	if result == "" {
		logs.GetLogger().Error("no response from:", cmdDeal.LotusClientApiUrl)
		return false, err
	}

	if string(result) == "0" {
		return false, nil
	}
	return true, nil
}

func minerFIdsIsExist(target, src []string) bool {
	var exist bool
loop:
	for _, t := range target {
		exist = false
		for _, s := range src {
			if strings.EqualFold(t, s) {
				exist = true
				break
			}
		}
		if !exist {
			break loop
		}
	}
	return exist
}
