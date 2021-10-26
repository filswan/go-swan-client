package subcommand

import (
	"fmt"
	"strings"

	"github.com/filswan/go-swan-client/model"
	"github.com/filswan/go-swan-lib/logs"

	"github.com/filswan/go-swan-lib/client"
	"github.com/filswan/go-swan-lib/constants"
	libmodel "github.com/filswan/go-swan-lib/model"
	"github.com/filswan/go-swan-lib/utils"
)

func SendAutoBidDeal(confDeal *model.ConfDeal) ([]string, error) {
	err := CreateOutputDir(confDeal.OutputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	logs.GetLogger().Info("output dir is:", confDeal.OutputDir)

	swanClient, err := client.SwanGetClient(confDeal.SwanApiUrl, confDeal.SwanApiKey, confDeal.SwanAccessToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	assignedTasks, err := swanClient.SwanGetAssignedTasks()
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
		assignedTaskInfo, err := swanClient.SwanGetOfflineDealsByTaskUuid(*assignedTask.Uuid)
		if err != nil {
			logs.GetLogger().Error(err)
			continue
		}

		deals := assignedTaskInfo.Data.Deal
		task := assignedTaskInfo.Data.Task
		dealSentNum, csvFilePath, err := SendAutobidDeal(confDeal, deals, task, confDeal.OutputDir)
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

		response, err := swanClient.SwanUpdateAssignedTask(*assignedTask.Uuid, status, csvFilePath)
		if err != nil {
			logs.GetLogger().Error(err)
			continue
		}

		logs.GetLogger().Info(response.Message)
	}

	return csvFilepaths, nil
}

func SendAutobidDeal(confDeal *model.ConfDeal, deals []libmodel.OfflineDeal, task libmodel.Task, outputDir string) (int, string, error) {
	carFiles := []*libmodel.FileDesc{}

	dealSentNum := 0
	for _, deal := range deals {
		err := model.SetDealConfig4Autobid(confDeal, task)
		if err != nil {
			logs.GetLogger().Error(err)
			continue
		}

		err = CheckDealConfig(confDeal)
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
		logs.GetLogger().Info("dealConfig.MinerPrice:", confDeal.MinerPrice)
		cost := CalculateRealCost(sectorSize, confDeal.MinerPrice)
		carFile := libmodel.FileDesc{
			Uuid:       task.Uuid,
			MinerFid:   task.MinerFid,
			CarFileUrl: deal.FileSourceUrl,
			CarFileMd5: deal.Md5Origin,
			StartEpoch: *deal.StartEpoch,
			PieceCid:   *deal.PieceCid,
			DataCid:    *deal.PayloadCid,
		}
		if carFile.MinerFid != nil {
			logs.GetLogger().Info("MinerFid:", *carFile.MinerFid)
		}

		logs.GetLogger().Info("FileSourceUrl:", carFile.CarFileUrl)
		carFiles = append(carFiles, &carFile)
		for i := 0; i < 60; i++ {
			msg := fmt.Sprintf("send deal for task:%s, deal:%d", task.TaskName, deal.Id)
			logs.GetLogger().Info(msg)
			dealConfig := libmodel.GetDealConfig(confDeal.VerifiedDeal, confDeal.FastRetrieval, confDeal.SkipConfirmation, confDeal.MinerPrice, confDeal.StartEpoch, *confDeal.MinerFid, confDeal.SenderWallet)
			dealCid, startEpoch, err := client.LotusProposeOfflineDeal(carFile, cost, pieceSize, *dealConfig, i)
			if err != nil {
				logs.GetLogger().Error(err)

				if strings.Contains(err.Error(), "already tracking identifier") {
					continue
				} else {
					break
				}
			}
			if dealCid == nil {
				continue
			}

			carFile.DealCid = dealCid
			carFile.StartEpoch = *startEpoch
			dealSentNum = dealSentNum + 1
			break
		}
	}

	jsonFileName := task.TaskName + "-autodeal-" + constants.JSON_FILE_NAME_BY_AUTO
	csvFileName := task.TaskName + "-autodeal-" + constants.CSV_FILE_NAME_BY_AUTO
	WriteCarFilesToFiles(carFiles, outputDir, jsonFileName, csvFileName)

	csvFilename := task.TaskName + "_autodeal.csv"
	csvFilepath, err := CreateCsv4TaskDeal(carFiles, outputDir, csvFilename)

	return dealSentNum, csvFilepath, err
}
