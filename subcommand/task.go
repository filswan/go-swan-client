package subcommand

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go-swan-client/common/client"
	"go-swan-client/common/constants"
	"go-swan-client/common/utils"
	"go-swan-client/config"
	"go-swan-client/logs"
	"go-swan-client/model"

	"github.com/google/uuid"
)

func CreateTask(inputDir string, taskName, outputDir, minerFid, dataset, description *string) (*string, error) {
	if len(inputDir) == 0 {
		err := fmt.Errorf("please provide input dir")
		logs.GetLogger().Error(err)
		return nil, err
	}

	if utils.GetPathType(inputDir) != constants.PATH_TYPE_DIR {
		err := fmt.Errorf("%s is not a directory", inputDir)
		logs.GetLogger().Error(err)
		return nil, err
	}

	outputDir, err := CreateOutputDir(outputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if taskName == nil || len(*taskName) == 0 {
		nowStr := "task_" + time.Now().Format("2006-01-02_15:04:05")
		taskName = &nowStr
	}

	publicDeal := config.GetConfig().Sender.PublicDeal
	verifiedDeal := config.GetConfig().Sender.VerifiedDeal
	offlineMode := config.GetConfig().Sender.OfflineMode
	fastRetrieval := config.GetConfig().Sender.FastRetrieval

	maxPrice := config.GetConfig().Sender.MaxPrice
	bidMode := config.GetConfig().Sender.BidMode
	startEpochHours := config.GetConfig().Sender.StartEpochHours
	expireDays := config.GetConfig().Sender.ExpireDays
	//generateMd5 := config.GetConfig().Sender.GenerateMd5

	storageServerType := config.GetConfig().Main.StorageServerType

	host := config.GetConfig().WebServer.Host
	port := config.GetConfig().WebServer.Port
	path := config.GetConfig().WebServer.Path

	downloadUrlPrefix := strings.TrimRight(host, "/") + ":" + strconv.Itoa(port)
	downloadUrlPrefix = filepath.Join(downloadUrlPrefix, path)

	logs.GetLogger().Info("swan client settings:")
	logs.GetLogger().Info("public task: ", publicDeal)
	logs.GetLogger().Info("verified deals: ", verifiedDeal)
	logs.GetLogger().Info("connected to swan: ", !offlineMode)
	logs.GetLogger().Info("csv/car file output dir: %s", outputDir)
	logs.GetLogger().Info("fastRetrieval: ", fastRetrieval)

	if !publicDeal && (minerFid == nil || len(*minerFid) == 0) {
		err := fmt.Errorf("please provide -miner for non public deal")
		logs.GetLogger().Error(err)
		return nil, err
	}

	carFiles := ReadCarFilesFromJsonFile(inputDir, constants.JSON_FILE_NAME_BY_UPLOAD)
	if carFiles == nil {
		err := fmt.Errorf("failed to read car files from :%s", inputDir)
		logs.GetLogger().Error(err)
		return nil, err
	}

	task := model.Task{
		TaskName:          *taskName,
		CuratedDataset:    *dataset,
		Description:       *description,
		FastRetrievalBool: fastRetrieval,
		MaxPrice:          &maxPrice,
		BidMode:           &bidMode,
		ExpireDays:        &expireDays,
		MinerFid:          minerFid,
		Uuid:              uuid.NewString(),
	}
	if publicDeal {
		task.IsPublic = 1
	} else {
		task.IsPublic = 0
	}

	if verifiedDeal {
		taskType := constants.TASK_TYPE_VERIFIED
		task.Type = &taskType
	} else {
		taskType := constants.TASK_TYPE_REGULAR
		task.Type = &taskType
	}

	for _, carFile := range carFiles {
		carFile.Uuid = task.Uuid
		carFile.StartEpoch = utils.GetCurrentEpoch() + (startEpochHours+1)*constants.EPOCH_PER_HOUR

		if storageServerType == constants.STORAGE_SERVER_TYPE_WEB_SERVER {
			carFile.CarFileUrl = filepath.Join(downloadUrlPrefix, carFile.CarFileName)
		}
	}

	if !publicDeal {
		_, err := SendDeals2Miner(nil, *taskName, *minerFid, *outputDir, carFiles)
		if err != nil {
			return nil, err
		}
	}

	jsonFileName := *taskName + constants.JSON_FILE_NAME_BY_TASK
	csvFileName := *taskName + constants.CSV_FILE_NAME_BY_TASK
	err = WriteCarFilesToFiles(carFiles, *outputDir, jsonFileName, csvFileName)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	err = SendTask2Swan(task, carFiles, *outputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return &jsonFileName, nil
}

func SendTask2Swan(task model.Task, carFiles []*model.FileDesc, outDir string) error {
	csvFilename := task.TaskName + "_task.csv"
	csvFilePath, err := CreateCsv4TaskDeal(carFiles, task.MinerFid, outDir, csvFilename)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	if config.GetConfig().Sender.OfflineMode {
		logs.GetLogger().Info("Working in Offline Mode. You need to manually send out task on filwan.com.")
		return nil
	}

	logs.GetLogger().Info("Working in Online Mode. A swan task will be created on the filwan.com after process done. ")

	swanClient := client.SwanGetClient()
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
