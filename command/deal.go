package command

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/filswan/go-swan-client/config"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/filswan/go-swan-lib/client/lotus"
	"github.com/filswan/go-swan-lib/client/swan"
	libconstants "github.com/filswan/go-swan-lib/constants"
	"github.com/filswan/go-swan-lib/logs"
	libmodel "github.com/filswan/go-swan-lib/model"
	"github.com/filswan/go-swan-lib/utils"
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
	StartDealTimeInterval  time.Duration   //required
}

func GetCmdDeal(outputDir *string, minerFids string, metadataJsonPath string) *CmdDeal {
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
		StartDealTimeInterval:  config.GetConfig().Sender.StartDealTimeInterval,
	}

	minerFids = strings.Trim(minerFids, " ")
	if minerFids != "" {
		cmdDeal.MinerFids = strings.Split(minerFids, ",")
	}

	if !utils.IsStrEmpty(outputDir) {
		cmdDeal.OutputDir = *outputDir
	} else {
		cmdDeal.OutputDir = filepath.Join(config.GetConfig().Sender.OutputDir, time.Now().Format("2006-01-02_15:04:05")) + "_" + uuid.NewString()
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

func SendDealsByConfig(outputDir, minerFid, metadataJsonPath string) ([]*libmodel.FileDesc, error) {
	if metadataJsonPath == "" {
		err := fmt.Errorf("metadataJsonPath is nil")
		logs.GetLogger().Error(err)
		return nil, err
	}

	cmdDeal := GetCmdDeal(&outputDir, minerFid, metadataJsonPath)
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

	fileDescs, err := ReadFileDescsFromJsonFileByFullPath(cmdDeal.MetadataJsonPath)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if len(fileDescs) == 0 {
		err := fmt.Errorf("no car files read from:%s", cmdDeal.MetadataJsonPath)
		logs.GetLogger().Error(err)
		return nil, err
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
		isWalletVerified, err := swanClient.CheckDatacap(cmdDeal.SenderWallet)
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
		}

		logs.GetLogger().Info("File:", fileDesc.CarFilePath, ",current epoch:", *currentEpoch, ", start epoch:", dealConfig.StartEpoch)

		if len(cmdDeal.MinerFids) == 0 {
			cmdDeal.MinerFids = []string{}
			for _, deal := range fileDesc.Deals {
				cmdDeal.MinerFids = append(cmdDeal.MinerFids, deal.MinerFid)
			}
		}

		if len(cmdDeal.MinerFids) == 0 {
			err := fmt.Errorf("miner is required, you can set in command line or in metadata json file")
			logs.GetLogger().Error(err)
			return nil, err
		}

		deals := []*libmodel.DealInfo{}
		for _, minerFid := range cmdDeal.MinerFids {
			dealConfig.MinerFid = minerFid

			dealCid, err := lotusClient.LotusClientStartDeal(&dealConfig)
			if dealCid == nil {
				continue
				dealCid = new(string)
			}
			deal := &libmodel.DealInfo{
				MinerFid:   dealConfig.MinerFid,
				DealCid:    *dealCid,
				StartEpoch: int(dealConfig.StartEpoch),
			}
			deals = append(deals, deal)
			dealSentNum = dealSentNum + 1
			if err != nil {
				logs.GetLogger().Error(err)
				continue
			}
			if dealCid == nil {
				continue
			}
			logs.GetLogger().Info("deal sent successfully, task name:", taskName, ", car file:", fileDesc.CarFilePath, ", deal CID:", deal.DealCid, ", start epoch:", deal.StartEpoch, ", miner:", deal.MinerFid)
			if cmdDeal.StartDealTimeInterval > 0 {
				time.Sleep(cmdDeal.StartDealTimeInterval * time.Millisecond)
			}
		}

		fileDesc.Deals = deals
	}

	logs.GetLogger().Info(dealSentNum, " deal(s) has(ve) been sent for task:", taskName)

	jsonFileName := taskName + JSON_FILE_NAME_DEAL
	_, err = WriteFileDescsToJsonFile(fileDescs, outputDir, jsonFileName)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return fileDescs, err
}
