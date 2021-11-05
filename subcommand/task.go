package subcommand

import (
	"fmt"

	"swan-client/model"

	"github.com/codingsince1985/checksum"
	"github.com/filswan/go-swan-lib/logs"

	"github.com/filswan/go-swan-lib/client/swan"
	"github.com/filswan/go-swan-lib/constants"
	libmodel "github.com/filswan/go-swan-lib/model"
	"github.com/filswan/go-swan-lib/utils"

	"github.com/google/uuid"
)

func CreateTask(confTask *model.ConfTask, confDeal *model.ConfDeal) (*string, []*libmodel.FileDesc, error) {
	if confTask == nil {
		err := fmt.Errorf("parameter confTask is nil")
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	if !confTask.PublicDeal {
		if confDeal == nil {
			err := fmt.Errorf("parameter confDeal is nil")
			logs.GetLogger().Error(err)
			return nil, nil, err
		}
	}

	err := CheckInputDir(confTask.InputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	err = CreateOutputDir(confTask.OutputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	logs.GetLogger().Info("you output dir: ", confTask.OutputDir)

	if !confTask.PublicDeal && len(confTask.MinerFid) == 0 {
		err := fmt.Errorf("please provide -miner for private deal")
		logs.GetLogger().Error(err)
		return nil, nil, err
	}
	if confTask.BidMode == constants.TASK_BID_MODE_AUTO && len(confTask.MinerFid) != 0 {
		logs.GetLogger().Warn("miner is unnecessary for aubo-bid task, it will be ignored")
	}

	if len(confTask.TaskName) == 0 {
		taskName := GetDefaultTaskName()
		confTask.TaskName = taskName
	}

	logs.GetLogger().Info("task settings:")
	logs.GetLogger().Info("public task: ", confTask.PublicDeal)
	logs.GetLogger().Info("verified deals: ", confTask.VerifiedDeal)
	logs.GetLogger().Info("connected to swan: ", !confTask.OfflineMode)
	logs.GetLogger().Info("fastRetrieval: ", confTask.FastRetrieval)

	carFiles := ReadCarFilesFromJsonFile(confTask.InputDir, constants.JSON_FILE_NAME_BY_UPLOAD)
	if carFiles == nil {
		err := fmt.Errorf("failed to read car files from :%s", confTask.InputDir)
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	isPublic := 0
	if confTask.PublicDeal {
		isPublic = 1
	}

	taskType := constants.TASK_TYPE_REGULAR
	if confTask.VerifiedDeal {
		taskType = constants.TASK_TYPE_VERIFIED
	}

	if confTask.Duration == 0 {
		confTask.Duration = DURATION
	}

	uuid := uuid.NewString()
	task := libmodel.Task{
		TaskName:          confTask.TaskName,
		FastRetrievalBool: confTask.FastRetrieval,
		Type:              taskType,
		IsPublic:          &isPublic,
		MaxPrice:          &confTask.MaxPrice,
		BidMode:           &confTask.BidMode,
		ExpireDays:        &confTask.ExpireDays,
		Uuid:              uuid,
		SourceId:          confTask.SourceId,
		Duration:          confTask.Duration,
		MinerFid:          confTask.MinerFid,
		CuratedDataset:    confTask.Dataset,
		Description:       confTask.Description,
	}

	for _, carFile := range carFiles {
		carFile.Uuid = task.Uuid
		carFile.MinerFid = task.MinerFid
		carFile.StartEpoch = &confTask.StartEpoch
		carFile.SourceId = &confTask.SourceId

		if confTask.StorageServerType == constants.STORAGE_SERVER_TYPE_WEB_SERVER {
			carFileUrl := utils.UrlJoin(confTask.WebServerDownloadUrlPrefix, carFile.CarFileName)
			carFile.CarFileUrl = carFileUrl
		}

		if confTask.GenerateMd5 {
			if carFile.SourceFileMd5 == "" {
				srcFileMd5, err := checksum.MD5sum(carFile.SourceFilePath)
				if err != nil {
					logs.GetLogger().Error(err)
					return nil, nil, err
				}
				carFile.SourceFileMd5 = srcFileMd5
			}

			if carFile.CarFileMd5 == "" {
				carFileMd5, err := checksum.MD5sum(carFile.CarFilePath)
				if err != nil {
					logs.GetLogger().Error(err)
					return nil, nil, err
				}
				carFile.CarFileMd5 = carFileMd5
			}
		}
	}

	if !confTask.PublicDeal {
		_, _, err := SendDeals2Miner(confDeal, confTask.TaskName, confTask.OutputDir, carFiles)
		if err != nil {
			return nil, nil, err
		}
	}

	jsonFileName := confTask.TaskName + constants.JSON_FILE_NAME_BY_TASK
	csvFileName := confTask.TaskName + constants.CSV_FILE_NAME_BY_TASK
	jsonFilepath, err := WriteCarFilesToFiles(carFiles, confTask.OutputDir, jsonFileName, csvFileName)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	err = SendTask2Swan(confTask, task, carFiles)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	return jsonFilepath, carFiles, nil
}

func SendTask2Swan(confTask *model.ConfTask, task libmodel.Task, carFiles []*libmodel.FileDesc) error {
	csvFilename := task.TaskName + ".csv"
	csvFilePath, err := CreateCsv4TaskDeal(carFiles, confTask.OutputDir, csvFilename)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	if confTask.OfflineMode {
		logs.GetLogger().Info("Working in Offline Mode. You need to manually send out task on filwan.com.")
		return nil
	}

	logs.GetLogger().Info("Working in Online Mode. A swan task will be created on the filwan.com after process done. ")
	swanClient, err := swan.SwanGetClient(confTask.SwanApiUrl, confTask.SwanApiKey, confTask.SwanAccessToken, confTask.SwanJwtToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	swanCreateTaskResponse, err := swanClient.SwanCreateTask(task, csvFilePath)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	if swanCreateTaskResponse.Status != "success" {
		err := fmt.Errorf("error, status%s, message:%s", swanCreateTaskResponse.Status, swanCreateTaskResponse.Message)
		logs.GetLogger().Info(err)
		return err
	}

	logs.GetLogger().Info("status:", swanCreateTaskResponse.Status, ", message:", swanCreateTaskResponse.Message)

	return nil
}
