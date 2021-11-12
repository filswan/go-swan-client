package subcommand

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/filswan/go-swan-client/model"

	"github.com/filswan/go-swan-lib/client/lotus"
	"github.com/filswan/go-swan-lib/client/swan"
	"github.com/filswan/go-swan-lib/constants"
	"github.com/filswan/go-swan-lib/logs"
	libmodel "github.com/filswan/go-swan-lib/model"
	"github.com/filswan/go-swan-lib/utils"
)

func SendDeals(confDeal *model.ConfDeal) ([]*libmodel.FileDesc, error) {
	if confDeal == nil {
		err := fmt.Errorf("parameter confDeal is nil")
		logs.GetLogger().Error(err)
		return nil, err
	}

	logs.GetLogger().Info(confDeal.OutputDir)
	err := CreateOutputDir(confDeal.OutputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	metadataJsonFilename := filepath.Base(confDeal.MetadataJsonPath)
	taskName := strings.TrimSuffix(metadataJsonFilename, constants.JSON_FILE_NAME_BY_TASK)
	carFiles := ReadCarFilesFromJsonFileByFullPath(confDeal.MetadataJsonPath)
	if len(carFiles) == 0 {
		err := fmt.Errorf("no car files read from:%s", confDeal.MetadataJsonPath)
		logs.GetLogger().Error(err)
		return nil, err
	}

	swanClient, err := swan.SwanGetClient(confDeal.SwanApiUrl, confDeal.SwanApiKey, confDeal.SwanAccessToken, confDeal.SwanToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	task, err := swanClient.SwanGetOfflineDealsByTaskUuid(carFiles[0].Uuid)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if task.Data.Task.IsPublic == nil || *task.Data.Task.IsPublic != constants.TASK_IS_PUBLIC {
		err := fmt.Errorf("task:%s is not in public mode,please check", task.Data.Task.TaskName)
		logs.GetLogger().Error(err)
		return nil, err
	}

	if task.Data.Task.BidMode == nil || *task.Data.Task.BidMode != constants.TASK_BID_MODE_MANUAL {
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
			err := fmt.Errorf("task:%s is verified, but your wallet:%s is not verified", taskName, confDeal.SenderWallet)
			logs.GetLogger().Error(err)
			return nil, err
		}
	}

	csvFilepath, carFiles, err := SendDeals2Miner(confDeal, taskName, confDeal.OutputDir, carFiles)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	err = swanClient.SwanUpdateTaskByUuid(carFiles[0].Uuid, confDeal.MinerFid, *csvFilepath)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return carFiles, nil
}

func SendDeals2Miner(confDeal *model.ConfDeal, taskName string, outputDir string, carFiles []*libmodel.FileDesc) (*string, []*libmodel.FileDesc, error) {
	if confDeal == nil {
		err := fmt.Errorf("parameter confDeal is nil")
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	err := CheckDealConfig(confDeal)
	if err != nil {
		err := errors.New("failed to pass deal config check")
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	dealSentNum := 0
	for _, carFile := range carFiles {
		if carFile.CarFileSize <= 0 {
			logs.GetLogger().Error("File:" + carFile.CarFilePath + " %s is too small")
			continue
		}
		pieceSize, sectorSize := utils.CalculatePieceSize(carFile.CarFileSize)
		logs.GetLogger().Info("dealConfig.MinerPrice:", confDeal.MinerPrice)
		cost := utils.CalculateRealCost(sectorSize, confDeal.MinerPrice)
		dealConfig := libmodel.GetDealConfig(confDeal.VerifiedDeal, confDeal.FastRetrieval, confDeal.SkipConfirmation, confDeal.MinerPrice, confDeal.StartEpoch, confDeal.Duration, confDeal.MinerFid, confDeal.SenderWallet)

		lotusClient, err := lotus.LotusGetClient(confDeal.LotusClientApiUrl, confDeal.LotusClientAccessToken)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, nil, err
		}

		dealCid, startEpoch, err := lotusClient.LotusClientStartDeal(*carFile, cost, pieceSize, *dealConfig, 0)
		//dealCid, err := client.LotusClientStartDeal(*carFile, cost, pieceSize, *dealConfig)
		if err != nil {
			logs.GetLogger().Error(err)
			continue
		}
		if dealCid == nil {
			continue
		}
		carFile.MinerFid = confDeal.MinerFid
		carFile.DealCid = *dealCid
		carFile.StartEpoch = startEpoch
		carFile.Price = GetDealPrice(cost, confDeal.Duration)

		dealSentNum = dealSentNum + 1
		logs.GetLogger().Info("task:", taskName, ", deal CID:", carFile.DealCid, ", start epoch:", *carFile.StartEpoch, ", deal sent to ", confDeal.MinerFid, " successfully")
	}

	logs.GetLogger().Info(dealSentNum, " deal(s) has(ve) been sent for task:", taskName)

	jsonFileName := taskName + constants.JSON_FILE_NAME_BY_DEAL
	csvFileName := taskName + constants.CSV_FILE_NAME_BY_DEAL
	_, err = WriteCarFilesToFiles(carFiles, outputDir, jsonFileName, csvFileName)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	csvFilename := taskName + ".csv"
	csvFilepath, err := CreateCsv4TaskDeal(carFiles, outputDir, csvFilename)

	return &csvFilepath, carFiles, err
}
