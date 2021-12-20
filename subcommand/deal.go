package subcommand

import (
	"errors"
	"fmt"
	"strings"

	"github.com/filswan/go-swan-client/model"

	"github.com/filswan/go-swan-client/common/constants"
	"github.com/filswan/go-swan-lib/client/lotus"
	"github.com/filswan/go-swan-lib/client/swan"
	libconstants "github.com/filswan/go-swan-lib/constants"
	"github.com/filswan/go-swan-lib/logs"
	libmodel "github.com/filswan/go-swan-lib/model"
	"github.com/filswan/go-swan-lib/utils"
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

	swanClient, err := swan.GetClient(confDeal.SwanApiUrlToken, confDeal.SwanApiUrl, confDeal.SwanApiKey, confDeal.SwanAccessToken, confDeal.SwanToken)
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

	dealSentNum := 0
	for _, fileDesc := range fileDescs {
		if fileDesc.CarFileSize <= 0 {
			logs.GetLogger().Error("File:" + fileDesc.CarFilePath + " %s is too small")
			continue
		}

		if len(fileDesc.Deals) == 0 {
			if confDeal.MinerFid != "" {
				fileDesc.Deals = []*libmodel.DealInfo{}
				deal := &libmodel.DealInfo{
					MinerFid: confDeal.MinerFid,
				}
				fileDesc.Deals = append(fileDesc.Deals, deal)
			} else {
				err := fmt.Errorf("miner is required, you can set in command line or in metadata json file")
				logs.GetLogger().Error(err)
				return nil, err
			}
		}

		swanClient, err := swan.GetClient(confDeal.SwanApiUrlToken, confDeal.SwanApiUrl, confDeal.SwanApiKey, confDeal.SwanAccessToken, confDeal.SwanToken)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}

		carFileResult, err := swanClient.GetCarFileByUuidUrl(fileDesc.Uuid, fileDesc.CarFileUrl)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}

		OfflineDeals := carFileResult.OfflineDeals

		for _, deal := range fileDesc.Deals {
			for _, offlineDeal := range OfflineDeals {
				if strings.EqualFold(deal.MinerFid, offlineDeal.MinerFid) {
					err := fmt.Errorf("%s,has already been sent to miner:%s,deal CID:%s", fileDesc.CarFileUrl, deal.MinerFid, offlineDeal.DealCid)
					logs.GetLogger().Error(err)
					return nil, err
				}
			}
		}

		for _, deal := range fileDesc.Deals {
			pieceSize, sectorSize := utils.CalculatePieceSize(fileDesc.CarFileSize)
			cost := utils.CalculateRealCost(sectorSize, confDeal.MinerPrice)
			dealConfig := libmodel.GetDealConfig(confDeal.VerifiedDeal, confDeal.FastRetrieval, confDeal.SkipConfirmation, confDeal.MinerPrice, int(confDeal.StartEpoch), int(confDeal.Duration), deal.MinerFid, confDeal.SenderWallet)

			err := CheckDealConfig(confDeal, dealConfig)
			if err != nil {
				err := errors.New("failed to pass deal config check")
				logs.GetLogger().Error(err)
				return nil, err
			}

			lotusClient, err := lotus.LotusGetClient(confDeal.LotusClientApiUrl, confDeal.LotusClientAccessToken)
			if err != nil {
				logs.GetLogger().Error(err)
				return nil, err
			}

			dealCid, startEpoch, err := lotusClient.LotusClientStartDeal(fileDesc.PayloadCid, fileDesc.PieceCid, cost, pieceSize, *dealConfig, 0)
			if err != nil {
				logs.GetLogger().Error(err)
				continue
			}
			if dealCid == nil {
				continue
			}
			deal.MinerFid = confDeal.MinerFid
			deal.DealCid = *dealCid
			deal.StartEpoch = *startEpoch

			dealSentNum = dealSentNum + 1
			logs.GetLogger().Info("task:", taskName, ", deal CID:", deal.DealCid, ", start epoch:", *fileDesc.StartEpoch, ", deal sent to ", confDeal.MinerFid, " successfully")
		}

	}

	logs.GetLogger().Info(dealSentNum, " deal(s) has(ve) been sent for task:", taskName)

	jsonFileName := taskName + constants.JSON_FILE_NAME_DEAL
	_, err := WriteFileDescsToJsonFile(fileDescs, outputDir, jsonFileName)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return fileDescs, err
}
