package subcommand

import (
	"fmt"
	"strings"

	"github.com/filswan/go-swan-client/model"
	"github.com/filswan/go-swan-lib/logs"

	"github.com/filswan/go-swan-lib/client/lotus"
	"github.com/filswan/go-swan-lib/client/swan"
	"github.com/filswan/go-swan-lib/constants"
	libmodel "github.com/filswan/go-swan-lib/model"
	"github.com/filswan/go-swan-lib/utils"
)

func SendAutoBidDeals(confDeal *model.ConfDeal) ([]string, [][]*libmodel.FileDesc, error) {
	err := CreateOutputDir(confDeal.OutputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	logs.GetLogger().Info("output dir is:", confDeal.OutputDir)

	swanClient, err := swan.SwanGetClient(confDeal.SwanApiUrl, confDeal.SwanApiKey, confDeal.SwanAccessToken, confDeal.SwanJwtToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	assignedTasks, err := swanClient.SwanGetAssignedTasks()
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}
	logs.GetLogger().Info("autobid Swan task count:", len(assignedTasks))
	if len(assignedTasks) == 0 {
		logs.GetLogger().Info("no autobid task to be dealt with")
		return nil, nil, nil
	}

	var tasksDeals [][]*libmodel.FileDesc
	csvFilepaths := []string{}
	for _, assignedTask := range assignedTasks {
		_, csvFilePath, carFiles, err := SendAutoBidDealsByTaskUuid(confDeal, assignedTask.Uuid)
		if err != nil {
			logs.GetLogger().Error(err)
			continue
		}

		tasksDeals = append(tasksDeals, carFiles)
		csvFilepaths = append(csvFilepaths, csvFilePath)
	}

	return csvFilepaths, tasksDeals, nil
}

func SendAutoBidDealsByTaskUuid(confDeal *model.ConfDeal, taskUuid string) (int, string, []*libmodel.FileDesc, error) {
	err := CreateOutputDir(confDeal.OutputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return 0, "", nil, err
	}

	logs.GetLogger().Info("output dir is:", confDeal.OutputDir)

	swanClient, err := swan.SwanGetClient(confDeal.SwanApiUrl, confDeal.SwanApiKey, confDeal.SwanAccessToken, confDeal.SwanJwtToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return 0, "", nil, err
	}

	assignedTaskInfo, err := swanClient.SwanGetOfflineDealsByTaskUuid(taskUuid)
	if err != nil {
		logs.GetLogger().Error(err)
		return 0, "", nil, err
	}

	deals := assignedTaskInfo.Data.Deal
	task := assignedTaskInfo.Data.Task
	dealSentNum, csvFilePath, carFiles, err := SendAutobidDeals4Task(confDeal, deals, task, confDeal.OutputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return 0, "", nil, err
	}

	msg := fmt.Sprintf("%d deal(s) sent for task:%s", dealSentNum, task.TaskName)
	logs.GetLogger().Info(msg)

	if dealSentNum == 0 {
		err := fmt.Errorf("no deal sent for task:%s", task.TaskName)
		logs.GetLogger().Info(err)
		return 0, "", nil, err
	}

	status := constants.TASK_STATUS_DEAL_SENT
	if dealSentNum != len(deals) {
		status = constants.TASK_STATUS_PROGRESS_WITH_FAILURE
	}

	response, err := swanClient.SwanUpdateAssignedTask(taskUuid, status, csvFilePath)
	if err != nil {
		logs.GetLogger().Error(err)
		return 0, "", nil, err
	}

	logs.GetLogger().Info(response.Message)

	return dealSentNum, csvFilePath, carFiles, nil
}

func SendAutobidDeals4Task(confDeal *model.ConfDeal, deals []libmodel.OfflineDeal, task libmodel.Task, outputDir string) (int, string, []*libmodel.FileDesc, error) {
	carFiles := []*libmodel.FileDesc{}

	dealSentNum := 0
	for _, deal := range deals {
		deal.DealCid = strings.Trim(deal.DealCid, " ")
		if len(deal.DealCid) != 0 {
			dealSentNum = dealSentNum + 1
			continue
		}

		err := model.SetDealConfig4Autobid(confDeal, task, deal)
		if err != nil {
			logs.GetLogger().Error(err)
			continue
		}

		err = CheckDealConfig(confDeal)
		if err != nil {
			logs.GetLogger().Error(err)
			continue
		}

		fileSizeInt := utils.GetInt64FromStr(deal.FileSize)
		if fileSizeInt <= 0 {
			logs.GetLogger().Error("file is too small")
			continue
		}
		pieceSize, sectorSize := utils.CalculatePieceSize(fileSizeInt)
		logs.GetLogger().Info("dealConfig.MinerPrice:", confDeal.MinerPrice)
		cost := utils.CalculateRealCost(sectorSize, confDeal.MinerPrice)
		carFile := libmodel.FileDesc{
			Uuid:       task.Uuid,
			MinerFid:   task.MinerFid,
			CarFileUrl: deal.FileSourceUrl,
			CarFileMd5: deal.Md5Origin,
			StartEpoch: &deal.StartEpoch,
			PieceCid:   deal.PieceCid,
			DataCid:    deal.PayloadCid,
		}
		if carFile.MinerFid != "" {
			logs.GetLogger().Info("MinerFid:", carFile.MinerFid)
		}

		logs.GetLogger().Info("FileSourceUrl:", carFile.CarFileUrl)
		carFiles = append(carFiles, &carFile)
		for i := 0; i < 60; i++ {
			msg := fmt.Sprintf("send deal for task:%s, deal:%d", task.TaskName, deal.Id)
			logs.GetLogger().Info(msg)
			dealConfig := libmodel.GetDealConfig(confDeal.VerifiedDeal, confDeal.FastRetrieval, confDeal.SkipConfirmation, confDeal.MinerPrice, confDeal.StartEpoch, confDeal.Duration, confDeal.MinerFid, confDeal.SenderWallet)
			dealCid, startEpoch, err := lotus.LotusProposeOfflineDeal(carFile, cost, pieceSize, *dealConfig, i)
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

			carFile.DealCid = *dealCid
			carFile.StartEpoch = startEpoch
			dealSentNum = dealSentNum + 1
			break
		}
	}

	jsonFileName := task.TaskName + "-autodeal-" + constants.JSON_FILE_NAME_BY_AUTO
	csvFileName := task.TaskName + "-autodeal-" + constants.CSV_FILE_NAME_BY_AUTO
	WriteCarFilesToFiles(carFiles, outputDir, jsonFileName, csvFileName)

	csvFilename := task.TaskName + "_autodeal.csv"
	csvFilepath, err := CreateCsv4TaskDeal(carFiles, outputDir, csvFilename)

	return dealSentNum, csvFilepath, carFiles, err
}
