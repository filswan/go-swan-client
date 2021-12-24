package subcommand

import (
	"fmt"

	"github.com/filswan/go-swan-client/model"

	"github.com/filswan/go-swan-client/common/constants"
	"github.com/filswan/go-swan-lib/client/lotus"
	"github.com/filswan/go-swan-lib/client/swan"
	libconstants "github.com/filswan/go-swan-lib/constants"
	"github.com/filswan/go-swan-lib/logs"
	libmodel "github.com/filswan/go-swan-lib/model"
)

func SendDealsByConfig(outputDir, minerFid, metadataJsonPath string) ([]*libmodel.FileDesc, error) {
	if metadataJsonPath == "" {
		err := fmt.Errorf("metadataJsonPath is nil")
		logs.GetLogger().Error(err)
		return nil, err
	}

	confDeal := model.GetConfDeal(&outputDir, minerFid, metadataJsonPath)
	fileDescs, err := SendDeals(confDeal)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return fileDescs, nil
}

func SendDeals(confDeal *model.ConfDeal) ([]*libmodel.FileDesc, error) {
	if confDeal == nil {
		err := fmt.Errorf("parameter confDeal is nil")
		logs.GetLogger().Error(err)
		return nil, err
	}

	err := CreateOutputDir(confDeal.OutputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	fileDescs := ReadFileDescsFromJsonFileByFullPath(confDeal.MetadataJsonPath)
	if len(fileDescs) == 0 {
		err := fmt.Errorf("no car files read from:%s", confDeal.MetadataJsonPath)
		logs.GetLogger().Error(err)
		return nil, err
	}

	swanClient, err := swan.GetClient(confDeal.SwanApiUrl, confDeal.SwanApiKey, confDeal.SwanAccessToken, confDeal.SwanToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	task, err := swanClient.GetTaskByUuid(fileDescs[0].Uuid)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if task.Data.Task.IsPublic == nil || *task.Data.Task.IsPublic != libconstants.TASK_IS_PUBLIC {
		err := fmt.Errorf("task:%s,uuid::%s is not in public mode,please check", task.Data.Task.TaskName, task.Data.Task.Uuid)
		logs.GetLogger().Error(err)
		return nil, err
	}

	if task.Data.Task.BidMode == nil || *task.Data.Task.BidMode != libconstants.TASK_BID_MODE_MANUAL {
		err := fmt.Errorf("auto_bid mode for task:%s is not manual, please check", task.Data.Task.TaskName)
		logs.GetLogger().Error(err)
		return nil, err
	}

	if confDeal.VerifiedDeal {
		isWalletVerified, err := swanClient.CheckDatacap(confDeal.SenderWallet)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}

		if !isWalletVerified {
			err := fmt.Errorf("task:%s is verified, but your wallet:%s is not verified", task.Data.Task.TaskName, confDeal.SenderWallet)
			logs.GetLogger().Error(err)
			return nil, err
		}
	}

	fileDescs, err = SendDeals2Miner(confDeal, task.Data.Task.TaskName, confDeal.OutputDir, fileDescs)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	_, err = swanClient.UpdateTaskAfterSendDealByUuid(task.Data.Task, fileDescs)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return fileDescs, nil
}

func SendDeals2Miner(confDeal *model.ConfDeal, taskName string, outputDir string, fileDescs []*libmodel.FileDesc) ([]*libmodel.FileDesc, error) {
	if confDeal == nil {
		err := fmt.Errorf("parameter confDeal is nil")
		logs.GetLogger().Error(err)
		return nil, err
	}

	lotusClient, err := lotus.LotusGetClient(confDeal.LotusClientApiUrl, confDeal.LotusClientAccessToken)
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
		dealConfig := libmodel.DealConfig{
			VerifiedDeal:     confDeal.VerifiedDeal,
			FastRetrieval:    confDeal.FastRetrieval,
			SkipConfirmation: confDeal.SkipConfirmation,
			MaxPrice:         confDeal.MaxPrice,
			StartEpoch:       confDeal.StartEpoch,
			//MinerFid:         confDeal.MinerFid,
			SenderWallet: confDeal.SenderWallet,
			Duration:     int(confDeal.Duration),
			TransferType: libconstants.LOTUS_TRANSFER_TYPE_MANUAL,
			PayloadCid:   fileDesc.PayloadCid,
			PieceCid:     fileDesc.PieceCid,
			FileSize:     fileDesc.CarFileSize,
		}

		if len(confDeal.MinerFids) == 0 {
			confDeal.MinerFids = []string{}
			for _, deal := range fileDesc.Deals {
				confDeal.MinerFids = append(confDeal.MinerFids, deal.MinerFid)
			}
		}

		if len(confDeal.MinerFids) == 0 {
			err := fmt.Errorf("miner is required, you can set in command line or in metadata json file")
			logs.GetLogger().Error(err)
			return nil, err
		}

		deals := []*libmodel.DealInfo{}
		for _, minerFid := range confDeal.MinerFids {
			dealConfig.MinerFid = minerFid

			dealCid, startEpoch, err := lotusClient.LotusClientStartDeal(&dealConfig, 0)
			if err != nil {
				logs.GetLogger().Error(err)
				continue
			}
			if dealCid == nil {
				continue
			}

			deal := &libmodel.DealInfo{
				MinerFid:   dealConfig.MinerFid,
				DealCid:    *dealCid,
				StartEpoch: int(*startEpoch),
			}
			deals = append(deals, deal)
			dealSentNum = dealSentNum + 1
			logs.GetLogger().Info("deal sent successfully, task:", taskName, ", car file:", fileDesc.CarFilePath, ", deal CID:", deal.DealCid, ", start epoch:", deal.StartEpoch, ", miner:", deal.MinerFid)
		}

		fileDesc.Deals = deals
	}

	logs.GetLogger().Info(dealSentNum, " deal(s) has(ve) been sent for task:", taskName)

	jsonFileName := taskName + constants.JSON_FILE_NAME_DEAL
	_, err = WriteFileDescsToJsonFile(fileDescs, outputDir, jsonFileName)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return fileDescs, err
}
