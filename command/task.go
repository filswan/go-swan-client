package command

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/filswan/go-swan-client/config"
	"github.com/shopspring/decimal"

	"github.com/codingsince1985/checksum"
	"github.com/filswan/go-swan-lib/logs"

	"github.com/filswan/go-swan-lib/client/lotus"
	"github.com/filswan/go-swan-lib/client/swan"
	libconstants "github.com/filswan/go-swan-lib/constants"
	libmodel "github.com/filswan/go-swan-lib/model"
	"github.com/filswan/go-swan-lib/utils"

	"github.com/google/uuid"
)

type CmdTask struct {
	SwanApiUrl                 string          //required when OfflineMode is false
	SwanApiKey                 string          //required when OfflineMode is false and SwanJwtToken is not provided
	SwanAccessToken            string          //required when OfflineMode is false and SwanJwtToken is not provided
	SwanToken                  string          //required when OfflineMode is false and SwanApiKey & SwanAccessToken are not provided
	LotusClientApiUrl          string          //required
	BidMode                    int             //required
	VerifiedDeal               bool            //required
	OfflineMode                bool            //required
	FastRetrieval              bool            //required
	MaxPrice                   decimal.Decimal //required
	StorageServerType          string          //required
	WebServerDownloadUrlPrefix string          //required only when StorageServerType is web server
	ExpireDays                 int             //required
	GenerateMd5                bool            //required
	Duration                   int             //not necessary, when not provided use default value:1512000
	OutputDir                  string          //required
	InputDir                   string          //required
	TaskName                   string          //not necessary, when not provided use default value:swan_task_xxxxxx
	Dataset                    string          //not necessary
	Description                string          //not necessary
	StartEpochHours            int             //required
	SourceId                   int             //required
	MaxAutoBidCopyNumber       int             //required only for public autobid deal
}

func GetCmdTask(inputDir string, outputDir *string, taskName, dataset, description string) *CmdTask {
	cmdTask := &CmdTask{
		SwanApiUrl:                 config.GetConfig().Main.SwanApiUrl,
		SwanApiKey:                 config.GetConfig().Main.SwanApiKey,
		SwanAccessToken:            config.GetConfig().Main.SwanAccessToken,
		LotusClientApiUrl:          config.GetConfig().Lotus.ClientApiUrl,
		BidMode:                    config.GetConfig().Sender.BidMode,
		VerifiedDeal:               config.GetConfig().Sender.VerifiedDeal,
		OfflineMode:                config.GetConfig().Sender.OfflineMode,
		FastRetrieval:              config.GetConfig().Sender.FastRetrieval,
		StorageServerType:          config.GetConfig().Main.StorageServerType,
		WebServerDownloadUrlPrefix: config.GetConfig().WebServer.DownloadUrlPrefix,
		ExpireDays:                 config.GetConfig().Sender.ExpireDays,
		GenerateMd5:                config.GetConfig().Sender.GenerateMd5,
		Duration:                   config.GetConfig().Sender.Duration,
		OutputDir:                  filepath.Join(config.GetConfig().Sender.OutputDir, time.Now().Format("2006-01-02_15:04:05")),
		InputDir:                   inputDir,
		TaskName:                   taskName,
		Dataset:                    dataset,
		Description:                description,
		StartEpochHours:            config.GetConfig().Sender.StartEpochHours,
		SourceId:                   libconstants.TASK_SOURCE_ID_SWAN_CLIENT,
		MaxAutoBidCopyNumber:       config.GetConfig().Sender.MaxAutoBidCopyNumber,
	}

	if !utils.IsStrEmpty(outputDir) {
		cmdTask.OutputDir = *outputDir
	}

	var err error
	maxPrice := config.GetConfig().Sender.MaxPrice
	cmdTask.MaxPrice, err = decimal.NewFromString(maxPrice)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	return cmdTask
}

func CreateTaskByConfig(inputDir string, outputDir *string, taskName, minerFid, dataset, description string) (*string, []*libmodel.FileDesc, []*Deal, error) {
	cmdTask := GetCmdTask(inputDir, outputDir, taskName, dataset, description)
	cmdDeal := GetCmdDeal(outputDir, minerFid, "")
	jsonFileName, fileDescs, deals, err := cmdTask.CreateTask(cmdDeal)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, nil, err
	}
	logs.GetLogger().Info("Task information is in:", *jsonFileName)

	return jsonFileName, fileDescs, deals, nil
}

func (cmdTask *CmdTask) CreateTask(cmdDeal *CmdDeal) (*string, []*libmodel.FileDesc, []*Deal, error) {
	if cmdTask.BidMode == libconstants.TASK_BID_MODE_NONE {
		if cmdDeal == nil {
			err := fmt.Errorf("parameter PublicDeal is nil")
			logs.GetLogger().Error(err)
			return nil, nil, nil, err
		}
	}

	lotusClient, err := lotus.LotusGetClient(cmdTask.LotusClientApiUrl, "")
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, nil, err
	}

	err = utils.CheckDirExists(cmdTask.InputDir, DIR_NAME_INPUT)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, nil, err
	}

	err = utils.CreateDirIfNotExists(cmdTask.OutputDir, DIR_NAME_OUTPUT)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, nil, err
	}

	logs.GetLogger().Info("Your output dir: ", cmdTask.OutputDir)
	if len(cmdTask.TaskName) == 0 {
		taskName := utils.GetDefaultTaskName()
		cmdTask.TaskName = taskName
	}

	fileDescs, err := ReadFileDescsFromJsonFile(cmdTask.InputDir, JSON_FILE_NAME_CAR_UPLOAD)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, nil, err
	}

	if fileDescs == nil {
		err := fmt.Errorf("failed to read car files from :%s", cmdTask.InputDir)
		logs.GetLogger().Error(err)
		return nil, nil, nil, err
	}

	if len(cmdDeal.MinerFids) > 0 {
		if cmdTask.BidMode == libconstants.TASK_BID_MODE_AUTO {
			logs.GetLogger().Warn("miner fids is unnecessary for auto-bid task")
		} else if cmdTask.BidMode == libconstants.TASK_BID_MODE_MANUAL {
			logs.GetLogger().Warn("miner fids is unnecessary for manual-bid task")
		}
	}

	taskType := libconstants.TASK_TYPE_REGULAR
	if cmdTask.VerifiedDeal {
		taskType = libconstants.TASK_TYPE_VERIFIED
	}

	if cmdTask.Duration == 0 {
		cmdTask.Duration = libconstants.DURATION_DEFAULT
	}

	currentEpoch, err := lotusClient.LotusGetCurrentEpoch()
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, nil, err
	}

	startEpoch := *currentEpoch + int64((cmdTask.StartEpochHours+1)*libconstants.EPOCH_PER_HOUR)

	err = lotusClient.CheckDuration(cmdTask.Duration, startEpoch)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, nil, err
	}

	fastRetrieval := libconstants.TASK_FAST_RETRIEVAL_NO
	if cmdTask.FastRetrieval {
		fastRetrieval = libconstants.TASK_FAST_RETRIEVAL_YES
	}

	uuid := uuid.NewString()
	task := libmodel.Task{
		TaskName:             cmdTask.TaskName,
		FastRetrieval:        &fastRetrieval,
		Type:                 taskType,
		MaxPrice:             &cmdTask.MaxPrice,
		BidMode:              &cmdTask.BidMode,
		ExpireDays:           &cmdTask.ExpireDays,
		Uuid:                 uuid,
		SourceId:             cmdTask.SourceId,
		Duration:             cmdTask.Duration,
		CuratedDataset:       cmdTask.Dataset,
		Description:          cmdTask.Description,
		MaxAutoBidCopyNumber: cmdTask.MaxAutoBidCopyNumber,
	}

	for _, fileDesc := range fileDescs {
		fileDesc.Uuid = task.Uuid
		fileDesc.StartEpoch = &startEpoch
		fileDesc.SourceId = &cmdTask.SourceId

		if cmdTask.StorageServerType == libconstants.STORAGE_SERVER_TYPE_WEB_SERVER {
			fileDesc.CarFileUrl = utils.UrlJoin(cmdTask.WebServerDownloadUrlPrefix, fileDesc.CarFileName)
		}

		if cmdTask.GenerateMd5 {
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

	if cmdTask.BidMode == libconstants.TASK_BID_MODE_NONE {
		_, err := cmdDeal.sendDeals2Miner(cmdTask.TaskName, cmdTask.OutputDir, fileDescs)
		if err != nil {
			return nil, nil, nil, err
		}
	}

	jsonFileName := cmdTask.TaskName + JSON_FILE_NAME_TASK
	jsonFilepath, err := WriteFileDescsToJsonFile(fileDescs, cmdTask.OutputDir, jsonFileName)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, nil, err
	}

	deals, err := cmdTask.sendTask2Swan(task, fileDescs)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, nil, err
	}

	if *task.BidMode == libconstants.TASK_BID_MODE_MANUAL {
		logs.GetLogger().Info("task ", task.TaskName, " has been created, please send its deal(s) later using deal subcommand and ", *jsonFilepath)
	}

	return jsonFilepath, fileDescs, deals, nil
}

func (cmdTask *CmdTask) sendTask2Swan(task libmodel.Task, fileDescs []*libmodel.FileDesc) ([]*Deal, error) {
	deals, err := GetDeals(fileDescs)
	if err != nil {
		logs.GetLogger().Error(err)
		return deals, err
	}

	if cmdTask.OfflineMode {
		logs.GetLogger().Info("Working in Offline Mode. You need to manually send out task on filwan.com.")
		return deals, nil
	}

	logs.GetLogger().Info("Working in Online Mode. A swan task will be created on the filwan.com after process done. ")
	swanClient, err := swan.GetClient(cmdTask.SwanApiUrl, cmdTask.SwanApiKey, cmdTask.SwanAccessToken, cmdTask.SwanToken)
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
