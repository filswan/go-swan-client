package subcommand

import (
	"errors"

	"github.com/DoraNebula/go-swan-client/common/client"
	"github.com/DoraNebula/go-swan-client/common/utils"
	"github.com/DoraNebula/go-swan-client/config"
	"github.com/DoraNebula/go-swan-client/logs"
	"github.com/DoraNebula/go-swan-client/model"

	"github.com/shopspring/decimal"
)

func SendAutoBidDeal(outputDir *string) {
	swanClient := client.SwanGetClient()
	assignedTasks, err := swanClient.GetAssignedTasks()
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}

	for _, assignedTask := range assignedTasks {
		assignedTaskInfo, err1 := swanClient.GetOfflineDealsByTaskUuid(assignedTask.Uuid)
		if err1 != nil {
			logs.GetLogger().Error(err1)
			continue
		}

		csvFilePath, err := SendAutobidDeal(assignedTaskInfo.Data.Deal, assignedTaskInfo.Data.Miner, assignedTaskInfo.Data.Task, outputDir)

		if err != nil {
			logs.GetLogger().Error(err)
			continue
		}

		response := swanClient.UpdateAssignedTask(assignedTask.Uuid, csvFilePath)
		logs.GetLogger().Info(response)
	}
}

func SendAutobidDeal(deals []model.OfflineDeal, miner model.Miner, task model.Task, outputDir *string) (string, error) {
	carFiles := []*model.FileDesc{}

	for _, deal := range deals {
		dealConfig := GetDealConfig1(task, deal)
		CheckDealConfig1(dealConfig)
		fileSizeInt := utils.GetInt64FromStr(*deal.FileSize)
		if fileSizeInt <= 0 {
			logs.GetLogger().Error("file is too small")
			continue
		}
		pieceSize, sectorSize := CalculatePieceSize(fileSizeInt)
		logs.GetLogger().Info("dealConfig.MinerPrice:", dealConfig.MinerPrice)
		cost := CalculateRealCost(sectorSize, dealConfig.MinerPrice)
		carFile := model.FileDesc{
			StartEpoch: *deal.StartEpoch,
			PieceCid:   *deal.PieceCid,
			DataCid:    *deal.PayloadCid,
		}
		carFiles = append(carFiles, &carFile)
		dealCid := client.LotusProposeOfflineDeal(carFile, cost, pieceSize, *dealConfig)
		if dealCid == nil {
			err := errors.New("failed to propose offline deal")
			logs.GetLogger().Error(err)
			return "", err
		}
		carFile.MinerFid = task.MinerFid
		carFile.DealCid = *dealCid
	}

	if outputDir == nil {
		outDir := config.GetConfig().Sender.OutputDir
		outputDir = &outDir
	}

	jsonFileName := task.TaskName + JSON_FILE_NAME_BY_AUTO_SUFFIX
	csvFileName := task.TaskName + CSV_FILE_NAME_BY_AUTO_SUFFIX
	WriteCarFilesToFiles(carFiles, *outputDir, jsonFileName, csvFileName)

	csvFilename := task.TaskName + "_deal.csv"
	csvFilepath, err := CreateCsv4TaskDeal(carFiles, task.MinerFid, *outputDir, csvFilename)

	return csvFilepath, err
}

func GetDealConfig1(task model.Task, deal model.OfflineDeal) *model.DealConfig {
	dealConfig := model.DealConfig{
		MinerFid:           *task.MinerFid,
		SenderWallet:       config.GetConfig().Sender.Wallet,
		VerifiedDeal:       config.GetConfig().Sender.VerifiedDeal,
		FastRetrieval:      config.GetConfig().Sender.FastRetrieval,
		EpochIntervalHours: config.GetConfig().Sender.StartEpochHours,
		SkipConfirmation:   config.GetConfig().Sender.SkipConfirmation,
		StartEpochHours:    *deal.StartEpoch,
	}

	maxPrice, err := decimal.NewFromString(*task.MaxPrice)
	if err != nil {
		logs.GetLogger().Error("Failed to convert maxPrice(" + *task.MaxPrice + ") to decimal, MaxPrice:")
		return nil
	}
	dealConfig.MaxPrice = maxPrice

	return &dealConfig
}

func CheckDealConfig1(dealConfig *model.DealConfig) bool {
	minerPrice, minerVerifiedPrice, _, _ := client.LotusGetMinerConfig(dealConfig.MinerFid)

	if dealConfig.SenderWallet == "" {
		logs.GetLogger().Error("Sender.wallet should be set in config file.")
		return false
	}

	if dealConfig.VerifiedDeal {
		if minerVerifiedPrice == nil {
			return false
		}
		dealConfig.MinerPrice = *minerVerifiedPrice
		logs.GetLogger().Info("Miner price is:", *minerVerifiedPrice)
	} else {
		if minerPrice == nil {
			return false
		}
		dealConfig.MinerPrice = *minerPrice
		logs.GetLogger().Info("Miner price is:", *minerPrice)
	}

	logs.GetLogger().Info("Miner price is:", dealConfig.MinerPrice, " MaxPrice:", dealConfig.MaxPrice, " VerifiedDeal:", dealConfig.VerifiedDeal)
	priceCmp := dealConfig.MaxPrice.Cmp(dealConfig.MinerPrice)
	logs.GetLogger().Info("priceCmp:", priceCmp)
	if priceCmp < 0 {
		logs.GetLogger().Error("miner price is higher than deal max price")
		return false
	}

	logs.GetLogger().Info("Deal check passed.")

	return true
}
