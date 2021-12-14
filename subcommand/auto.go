package subcommand

import (
	"fmt"
	"strings"
	"time"

	"github.com/filswan/go-swan-client/model"

	"github.com/filswan/go-swan-lib/logs"

	"github.com/filswan/go-swan-client/common/constants"
	"github.com/filswan/go-swan-lib/client/lotus"
	"github.com/filswan/go-swan-lib/client/swan"
	libconstants "github.com/filswan/go-swan-lib/constants"
	libmodel "github.com/filswan/go-swan-lib/model"
	"github.com/filswan/go-swan-lib/utils"
)

func SendAutoBidDealsLoop(confDeal *model.ConfDeal) error {
	err := CreateOutputDir(confDeal.OutputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	for {
		_, err := SendAutoBidDeals(confDeal)
		if err != nil {
			logs.GetLogger().Error(err)
			//return err
			continue
		}

		time.Sleep(time.Second * 30)
	}
}

func SendAutoBidDeals(confDeal *model.ConfDeal) ([][]*libmodel.FileDesc, error) {
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

	logs.GetLogger().Info("output dir is:", confDeal.OutputDir)

	swanClient, err := swan.SwanGetClient(confDeal.SwanApiUrl, confDeal.SwanApiKey, confDeal.SwanAccessToken, confDeal.SwanToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	assignedTasks, err := swanClient.SwanGetAllTasks(libconstants.TASK_STATUS_ASSIGNED)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	logs.GetLogger().Info("autobid Swan task count:", len(assignedTasks))
	if len(assignedTasks) == 0 {
		logs.GetLogger().Info("no autobid task to be dealt with")
		return nil, nil
	}

	var tasksDeals [][]*libmodel.FileDesc
	for _, assignedTask := range assignedTasks {
		if !IsTaskSourceRight(confDeal, assignedTask) {
			continue
		}

		_, carFiles, err := SendAutoBidDealsByTaskUuid(confDeal, assignedTask.Uuid)
		if err != nil {
			logs.GetLogger().Error(err)
			continue
		}

		tasksDeals = append(tasksDeals, carFiles)
	}

	return tasksDeals, nil
}

func SendAutoBidDealsByTaskUuid(confDeal *model.ConfDeal, taskUuid string) (int, []*libmodel.FileDesc, error) {
	if confDeal == nil {
		err := fmt.Errorf("parameter confDeal is nil")
		logs.GetLogger().Error(err)
		return 0, nil, err
	}

	err := CreateOutputDir(confDeal.OutputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return 0, nil, err
	}

	logs.GetLogger().Info("output dir is:", confDeal.OutputDir)

	swanClient, err := swan.SwanGetClient(confDeal.SwanApiUrl, confDeal.SwanApiKey, confDeal.SwanAccessToken, confDeal.SwanToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return 0, nil, err
	}

	assignedTaskInfo, err := swanClient.SwanGetTaskByUuid(taskUuid)
	if err != nil {
		logs.GetLogger().Error(err)
		return 0, nil, err
	}

	carFiles := assignedTaskInfo.Data.Deal
	task := assignedTaskInfo.Data.Task

	if task.Type == libconstants.TASK_TYPE_VERIFIED {
		isWalletVerified, err := swanClient.CheckDatacap(confDeal.SenderWallet)
		if err != nil {
			logs.GetLogger().Error(err)
			return 0, nil, err
		}

		if !isWalletVerified {
			err := fmt.Errorf("task:%s is verified, but your wallet:%s is not verified", task.TaskName, confDeal.SenderWallet)
			logs.GetLogger().Error(err)
			return 0, nil, err
		}
	}

	dealNum, dealSentNum, carFiles, err := SendAutobidDeals4Task(confDeal, task, confDeal.OutputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return 0, nil, err
	}

	msg := fmt.Sprintf("%d deal(s) sent to:%s for task:%s", dealSentNum, confDeal.MinerFid, task.TaskName)
	logs.GetLogger().Info(msg)

	if dealSentNum == 0 {
		err := fmt.Errorf("no deal sent for task:%s", task.TaskName)
		logs.GetLogger().Info(err)
		return 0, nil, err
	}

	status := libconstants.TASK_STATUS_DEAL_SENT
	if dealSentNum != dealNum {
		status = libconstants.TASK_STATUS_PROGRESS_WITH_FAILURE
	}

	logs.GetLogger().Info(status)
	_, err = swanClient.SwanUpdateTaskByUuid(task, carFiles)
	if err != nil {
		logs.GetLogger().Error(err)
		return 0, nil, err
	}

	return dealSentNum, carFiles, nil
}
func SendAutobidDeals4Task(confDeal *model.ConfDeal, task libmodel.Task, carFiles []*model.CarFile, outputDir string) (int, int, []*libmodel.FileDesc, error) {
	fileDescs := []*libmodel.FileDesc{}
	allDealNum := 0
	allDealSentNum := 0
	for _, carFile := range carFiles {
		offlineDeals := []*libmodel.OfflineDeal{}
		allDealNum = allDealNum + len(offlineDeals)
		dealSentNum, fileDesc, err := SendAutobidDeals4CarFile(confDeal, offlineDeals, carFile, task, outputDir)
		if err != nil {
			logs.GetLogger().Error(err)
			return 0, 0, nil, err
		}
		fileDescs = append(fileDescs, fileDesc)

		allDealSentNum = allDealSentNum + dealSentNum
	}
	jsonFileName := task.TaskName + constants.JSON_FILE_NAME_DEAL_AUTO
	_, err := WriteCarFilesToJsonFile(fileDescs, outputDir, jsonFileName, SUBCOMMAND_AUTO)
	if err != nil {
		logs.GetLogger().Error(err)
		return 0, 0, nil, err
	}

	return allDealNum, allDealSentNum, fileDescs, nil
}
func SendAutobidDeals4CarFile(confDeal *model.ConfDeal, offlineDeals []*libmodel.OfflineDeal, carFile *model.CarFile, task libmodel.Task, outputDir string) (int, *libmodel.FileDesc, error) {
	if confDeal == nil {
		err := fmt.Errorf("parameter confDeal is nil")
		logs.GetLogger().Error(err)
		return 0, nil, err
	}

	if !IsTaskSourceRight(confDeal, task) {
		err := fmt.Errorf("you cannot send deal from this kind of source")
		logs.GetLogger().Error(err)
		return 0, nil, err
	}

	fileDesc := libmodel.FileDesc{
		Uuid:        task.Uuid,
		Deals:       []libmodel.DealInfo{},
		CarFileUrl:  carFile.FileUrl,
		CarFileMd5:  *carFile.FileMd5,
		PieceCid:    carFile.PieceCid,
		PayloadCid:  carFile.PayloadCid,
		CarFileSize: int64(carFile.FileSize),
	}
	dealSentNum := 0
	for _, offlineDeal := range offlineDeals {
		offlineDeal.DealCid = strings.Trim(offlineDeal.DealCid, " ")
		if len(offlineDeal.DealCid) != 0 {
			dealSentNum = dealSentNum + 1
			continue
		}

		err := model.SetDealConfig4Autobid(confDeal, task, *offlineDeal)
		if err != nil {
			logs.GetLogger().Error(err)
			continue
		}

		err = CheckDealConfig(confDeal)
		if err != nil {
			logs.GetLogger().Error(err)
			continue
		}

		fileSizeInt := utils.GetInt64FromStr(offlineDeal.FileSize)
		if fileSizeInt <= 0 {
			logs.GetLogger().Error("file is too small")
			continue
		}
		pieceSize, sectorSize := utils.CalculatePieceSize(fileSizeInt)
		cost := utils.CalculateRealCost(sectorSize, confDeal.MinerPrice)
		for i := 0; i < 60; i++ {
			msg := fmt.Sprintf("send deal for task:%s, deal:%d", task.TaskName, offlineDeal.Id)
			logs.GetLogger().Info(msg)
			dealConfig := libmodel.GetDealConfig(confDeal.VerifiedDeal, confDeal.FastRetrieval, confDeal.SkipConfirmation, confDeal.MinerPrice, confDeal.StartEpoch, confDeal.Duration, confDeal.MinerFid, confDeal.SenderWallet)

			lotusClient, err := lotus.LotusGetClient(confDeal.LotusClientApiUrl, confDeal.LotusClientAccessToken)
			if err != nil {
				logs.GetLogger().Error(err)
				return 0, nil, err
			}

			dealCid, startEpoch, err := lotusClient.LotusClientStartDeal(fileDesc, cost, pieceSize, *dealConfig, i)
			if err != nil {
				logs.GetLogger().Error("tried ", i, " times,", err)

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

			dealInfo := libmodel.DealInfo{
				MinerFid:   task.MinerFid,
				DealCid:    *dealCid,
				StartEpoch: *startEpoch,
			}
			fileDesc.Deals = append(fileDesc.Deals, dealInfo)
			dealSentNum = dealSentNum + 1

			logs.GetLogger().Info("task:", task.TaskName, ", deal CID:", dealInfo.DealCid, ", start epoch:", dealInfo.StartEpoch, ", deal sent to ", confDeal.MinerFid, " successfully")
			break
		}
	}
	return dealSentNum, &fileDesc, nil
}
