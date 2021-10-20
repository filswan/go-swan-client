package subcommand

import (
	"go-swan-client/logs"
	"go-swan-client/model"

	"go-swan-client/common/client"
	"go-swan-client/common/constants"
	"go-swan-client/common/utils"
	"go-swan-client/config"
)

func SendAutoBidDeal(outputDir *string) ([]string, error) {
	outputDir, err := CreateOutputDir(outputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	logs.GetLogger().Info("output dir is:", *outputDir)

	swanClient := client.SwanGetClient()
	assignedTasks, err := swanClient.GetAssignedTasks()
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	logs.GetLogger().Info("autobid Swan task count:", len(assignedTasks))
	if len(assignedTasks) == 0 {
		logs.GetLogger().Info("no autobid task to be dealt with")
		return nil, nil
	}

	csvFilepaths := []string{}
	for _, assignedTask := range assignedTasks {
		assignedTaskInfo, err := swanClient.GetOfflineDealsByTaskUuid(assignedTask.Uuid)
		if err != nil {
			logs.GetLogger().Error(err)
			continue
		}

		deals := assignedTaskInfo.Data.Deal
		miner := assignedTaskInfo.Data.Miner
		task := assignedTaskInfo.Data.Task
		dealSentNum, csvFilePath, err := SendAutobidDeal(deals, miner, task, outputDir)
		if err != nil {
			csvFilepaths = append(csvFilepaths, csvFilePath)
			logs.GetLogger().Error(err)
			continue
		}

		if dealSentNum == 0 {
			logs.GetLogger().Info(dealSentNum, " deal(s) sent for task:", task.TaskName)
			continue
		}

		status := constants.TASK_STATUS_DEAL_SENT
		if dealSentNum != len(deals) {
			status = constants.TASK_STATUS_PROGRESS_WITH_FAILURE
		}

		response, err := swanClient.UpdateAssignedTask(assignedTask.Uuid, status, csvFilePath)
		if err != nil {
			logs.GetLogger().Error(err)
			continue
		}

		logs.GetLogger().Info(response.Message)
	}

	return csvFilepaths, nil
}

func SendAutobidDeal(deals []model.OfflineDeal, miner model.Miner, task model.Task, outputDir *string) (int, string, error) {
	carFiles := []*model.FileDesc{}

	dealSentNum := 0
	for _, deal := range deals {
		dealConfig := GetDealConfig4Autobid(task, deal)
		err := CheckDealConfig(dealConfig)
		if err != nil {
			logs.GetLogger().Error(err)
			continue
		}
		fileSizeInt := utils.GetInt64FromStr(*deal.FileSize)
		if fileSizeInt <= 0 {
			logs.GetLogger().Error("file is too small")
			continue
		}
		pieceSize, sectorSize := CalculatePieceSize(fileSizeInt)
		logs.GetLogger().Info("dealConfig.MinerPrice:", dealConfig.MinerPrice)
		cost := CalculateRealCost(sectorSize, dealConfig.MinerPrice)
		carFile := model.FileDesc{
			Uuid:       task.Uuid,
			MinerFid:   task.MinerFid,
			CarFileUrl: *deal.FileSourceUrl,
			CarFileMd5: deal.Md5Origin,
			StartEpoch: *deal.StartEpoch,
			PieceCid:   *deal.PieceCid,
			DataCid:    *deal.PayloadCid,
		}
		logs.GetLogger().Info("FileSourceUrl:", carFile.CarFileUrl)
		carFiles = append(carFiles, &carFile)
		for i := 0; i < 60; i++ {
			dealCid, startEpoch, err := client.LotusProposeOfflineDeal(carFile, cost, pieceSize, *dealConfig, i)
			if err != nil {
				logs.GetLogger().Error(err)
				continue
			}
			if dealCid == nil {
				continue
			}

			carFile.DealCid = *dealCid
			carFile.StartEpoch = *startEpoch
			dealSentNum = dealSentNum + 1
			break
		}
	}

	if outputDir == nil {
		outDir := config.GetConfig().Sender.OutputDir
		outputDir = &outDir
	}

	jsonFileName := task.TaskName + "-autodeal-" + constants.JSON_FILE_NAME_BY_AUTO
	csvFileName := task.TaskName + "-autodeal-" + constants.CSV_FILE_NAME_BY_AUTO
	WriteCarFilesToFiles(carFiles, *outputDir, jsonFileName, csvFileName)

	csvFilename := task.TaskName + "_autodeal.csv"
	csvFilepath, err := CreateCsv4TaskDeal(carFiles, *outputDir, csvFilename)

	return dealSentNum, csvFilepath, err
}

func GetDealConfig4Autobid(task model.Task, deal model.OfflineDeal) *model.DealConfig {
	dealConfig := model.DealConfig{
		MinerFid:           *task.MinerFid,
		SenderWallet:       config.GetConfig().Sender.Wallet,
		VerifiedDeal:       *task.Type == constants.TASK_TYPE_VERIFIED,
		FastRetrieval:      *task.FastRetrieval == constants.TASK_FAST_RETRIEVAL,
		EpochIntervalHours: config.GetConfig().Sender.StartEpochHours,
		SkipConfirmation:   config.GetConfig().Sender.SkipConfirmation,
		StartEpochHours:    *deal.StartEpoch,
	}

	dealConfig.MaxPrice = *task.MaxPrice

	return &dealConfig
}
