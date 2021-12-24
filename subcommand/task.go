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

func CreateTaskByConfig(inputDir string, outputDir *string, taskName, minerFid, dataset, description string) (*string, []*libmodel.FileDesc, []*Deal, error) {
	confTask := model.GetConfTask(inputDir, outputDir, taskName, dataset, description)
	confDeal := model.GetConfDeal(outputDir, minerFid, "")
	jsonFileName, fileDescs, deals, err := CreateTask(confTask, confDeal)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, nil, err
	}
	logs.GetLogger().Info("Task information is in:", *jsonFileName)

	return jsonFileName, fileDescs, deals, nil
}

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
		taskName := utils.GetDefaultTaskName()
		confTask.TaskName = taskName
	}

	fileDescs := ReadFileDescsFromJsonFile(confTask.InputDir, constants.JSON_FILE_NAME_CAR_UPLOAD)
	if fileDescs == nil {
		err := fmt.Errorf("failed to read car files from :%s", confTask.InputDir)
		logs.GetLogger().Error(err)
		return nil, nil, nil, err
	}

	isPublic := 0
	if confTask.PublicDeal {
		isPublic = 1
		if len(confDeal.MinerFids) > 0 {
			logs.GetLogger().Warn("miner fids is unnecessary for public task")
		}
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

	fastRetrieval := libconstants.TASK_FAST_RETRIEVAL_NO
	if confTask.FastRetrieval {
		fastRetrieval = libconstants.TASK_FAST_RETRIEVAL_YES
	}

	uuid := uuid.NewString()
	task := libmodel.Task{
		TaskName:       confTask.TaskName,
		FastRetrieval:  &fastRetrieval,
		Type:           taskType,
		IsPublic:       &isPublic,
		MaxPrice:       &confTask.MaxPrice,
		BidMode:        &confTask.BidMode,
		ExpireDays:     &confTask.ExpireDays,
		Uuid:           uuid,
		SourceId:       confTask.SourceId,
		Duration:       confTask.Duration,
		CuratedDataset: confTask.Dataset,
		Description:    confTask.Description,
	}

	for _, fileDesc := range fileDescs {
		fileDesc.Uuid = task.Uuid
		fileDesc.StartEpoch = &confTask.StartEpoch
		fileDesc.SourceId = &confTask.SourceId

		if confTask.StorageServerType == libconstants.STORAGE_SERVER_TYPE_WEB_SERVER {
			fileDesc.CarFileUrl = utils.UrlJoin(confTask.WebServerDownloadUrlPrefix, fileDesc.CarFileName)
		}

		if confTask.GenerateMd5 {
			if fileDesc.SourceFileMd5 == "" && utils.IsFileExistsFullPath(fileDesc.SourceFilePath) {
				srcFileMd5, err := checksum.MD5sum(fileDesc.SourceFilePath)
				if err != nil {
					logs.GetLogger().Error(err)
					return nil, nil, nil, err
				}
				fileDesc.SourceFileMd5 = srcFileMd5
			}

			if fileDesc.CarFileMd5 == "" {
				carFileMd5, err := checksum.MD5sum(fileDesc.CarFilePath)
				if err != nil {
					logs.GetLogger().Error(err)
					return nil, nil, nil, err
				}
				fileDesc.CarFileMd5 = carFileMd5
			}
		}
	}

	if !confTask.PublicDeal {
		_, err := SendDeals2Miner(confDeal, confTask.TaskName, confTask.OutputDir, fileDescs)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	jsonFileName := confTask.TaskName + constants.JSON_FILE_NAME_TASK
	jsonFilepath, err := WriteFileDescsToJsonFile(fileDescs, confTask.OutputDir, jsonFileName)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, nil, err
	}

	deals, err := SendTask2Swan(confTask, task, fileDescs)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, nil, err
	}

	if *task.IsPublic == libconstants.TASK_IS_PUBLIC && *task.BidMode == libconstants.TASK_BID_MODE_MANUAL {
		logs.GetLogger().Info("task ", task.TaskName, " has been created, please send its deal(s) later using deal subcommand and ", *jsonFilepath)
	}

	return jsonFilepath, fileDescs, deals, nil
}

func SendTask2Swan(confTask *model.ConfTask, task libmodel.Task, fileDescs []*libmodel.FileDesc) ([]*Deal, error) {
	deals, err := GetDeals(fileDescs)
	if err != nil {
		logs.GetLogger().Error(err)
		return deals, err
	}

	if confTask.OfflineMode {
		logs.GetLogger().Info("Working in Offline Mode. You need to manually send out task on filwan.com.")
		return deals, nil
	}

	logs.GetLogger().Info("Working in Online Mode. A swan task will be created on the filwan.com after process done. ")
	swanClient, err := swan.GetClient(confTask.SwanApiUrl, confTask.SwanApiKey, confTask.SwanAccessToken, confTask.SwanToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return deals, err
	}

	swanCreateTaskResponse, err := swanClient.CreateTask(task, fileDescs)
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
