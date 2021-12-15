package subcommand

import (
	"fmt"

	"github.com/filswan/go-swan-client/common/constants"
	"github.com/filswan/go-swan-client/model"

	"github.com/codingsince1985/checksum"
	"github.com/filswan/go-swan-lib/logs"

	"github.com/filswan/go-swan-lib/client/swan"
	libconstants "github.com/filswan/go-swan-lib/constants"
	libmodel "github.com/filswan/go-swan-lib/model"
	"github.com/filswan/go-swan-lib/utils"

	"github.com/google/uuid"
)

func CreateTask(confTask *model.ConfTask, confDeal *model.ConfDeal) (*string, []*libmodel.FileDesc, []*Deal, error) {
	if confTask == nil {
		err := fmt.Errorf("parameter confTask is nil")
		logs.GetLogger().Error(err)
		return nil, nil, nil, err
	}

	if !confTask.PublicDeal {
		if confDeal == nil {
			err := fmt.Errorf("parameter confDeal is nil")
			logs.GetLogger().Error(err)
			return nil, nil, nil, err
		}
	}

	err := CheckInputDir(confTask.InputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, nil, err
	}

	err = CreateOutputDir(confTask.OutputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, nil, err
	}

	logs.GetLogger().Info("you output dir: ", confTask.OutputDir)
	if len(confTask.TaskName) == 0 {
		taskName := GetDefaultTaskName()
		confTask.TaskName = taskName
	}

	carFiles := ReadCarFilesFromJsonFile(confTask.InputDir, constants.JSON_FILE_NAME_CAR_UPLOAD)
	if carFiles == nil {
		err := fmt.Errorf("failed to read car files from :%s", confTask.InputDir)
		logs.GetLogger().Error(err)
		return nil, nil, nil, err
	}

	isPublic := 0
	if confTask.PublicDeal {
		isPublic = 1
	}

	taskType := libconstants.TASK_TYPE_REGULAR
	if confTask.VerifiedDeal {
		taskType = libconstants.TASK_TYPE_VERIFIED
	}

	if confTask.Duration == 0 {
		confTask.Duration = DURATION
	}

	err = CheckDuration(confTask.Duration, confTask.StartEpoch, 0)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, nil, err
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
		CuratedDataset:    confTask.Dataset,
		Description:       confTask.Description,
	}

	for _, carFile := range carFiles {
		carFile.Uuid = task.Uuid
		carFile.StartEpoch = &confTask.StartEpoch
		carFile.SourceId = &confTask.SourceId

		if confTask.StorageServerType == libconstants.STORAGE_SERVER_TYPE_WEB_SERVER {
			carFile.CarFileUrl = utils.UrlJoin(confTask.WebServerDownloadUrlPrefix, carFile.CarFileName)
		}

		if confTask.GenerateMd5 {
			if carFile.SourceFileMd5 == "" && utils.IsFileExistsFullPath(carFile.SourceFilePath) {
				srcFileMd5, err := checksum.MD5sum(carFile.SourceFilePath)
				if err != nil {
					logs.GetLogger().Error(err)
					return nil, nil, nil, err
				}
				carFile.SourceFileMd5 = srcFileMd5
			}

			if carFile.CarFileMd5 == "" {
				carFileMd5, err := checksum.MD5sum(carFile.CarFilePath)
				if err != nil {
					logs.GetLogger().Error(err)
					return nil, nil, nil, err
				}
				carFile.CarFileMd5 = carFileMd5
			}
		}
	}

	if !confTask.PublicDeal {
		_, err := SendDeals2Miner(confDeal, confTask.TaskName, confTask.OutputDir, carFiles)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	jsonFileName := confTask.TaskName + constants.JSON_FILE_NAME_TASK
	jsonFilepath, err := WriteCarFilesToJsonFile(carFiles, confTask.OutputDir, jsonFileName, SUBCOMMAND_TASK)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, nil, err
	}

	deals, err := SendTask2Swan(confTask, task, carFiles)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, nil, err
	}

	if *task.IsPublic == libconstants.TASK_IS_PUBLIC && *task.BidMode == libconstants.TASK_BID_MODE_MANUAL {
		logs.GetLogger().Info("task ", task.TaskName, " has been created, please send its deal(s) later using deal subcommand and ", *jsonFilepath)
	}

	return jsonFilepath, carFiles, deals, nil
}

func SendTask2Swan(confTask *model.ConfTask, task libmodel.Task, carFiles []*libmodel.FileDesc) ([]*Deal, error) {
	deals, err := GetDeals(carFiles)
	if err != nil {
		logs.GetLogger().Error(err)
		return deals, err
	}

	if confTask.OfflineMode {
		logs.GetLogger().Info("Working in Offline Mode. You need to manually send out task on filwan.com.")
		return deals, nil
	}

	logs.GetLogger().Info("Working in Online Mode. A swan task will be created on the filwan.com after process done. ")
	swanClient, err := swan.SwanGetClient(confTask.SwanApiUrlToken, confTask.SwanApiUrl, confTask.SwanApiKey, confTask.SwanAccessToken, confTask.SwanToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return deals, err
	}

	swanCreateTaskResponse, err := swanClient.SwanCreateTask(task, carFiles)
	if err != nil {
		logs.GetLogger().Error(err)
		return deals, err
	}

	if swanCreateTaskResponse.Status != "success" {
		err := fmt.Errorf("error, status%s, message:%s", swanCreateTaskResponse.Status, swanCreateTaskResponse.Message)
		logs.GetLogger().Info(err)
		return deals, err
	}

	logs.GetLogger().Info("status:", swanCreateTaskResponse.Status, ", message:", swanCreateTaskResponse.Message)

	return deals, nil
}
