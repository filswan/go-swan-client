package subcommand

import (
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

func CreateTask(taskName, inputDir, outputDir, minerFid, dataset, description *string) (*string, bool) {
	if outputDir == nil || len(*outputDir) == 0 {
		outDir := config.GetConfig().Sender.OutputDir
		outputDir = &outDir
	}

	if taskName == nil || len(*taskName) == 0 {
		nowStr := "task_" + time.Now().Format("2006-01-02_15:04:05")
		taskName = &nowStr
	}
	publicDeal := config.GetConfig().Sender.PublicDeal

	verifiedDeal := config.GetConfig().Sender.VerifiedDeal
	offlineMode := config.GetConfig().Sender.OfflineMode
	fastRetrieval := 0
	if config.GetConfig().Sender.FastRetrieval {
		fastRetrieval = 1
	}
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

	logs.GetLogger().Info("Swan Client Settings: Public Task: ", publicDeal, ",  Verified Deals: ", verifiedDeal, ",  Connected to Swan: ", !offlineMode, ", CSV/car File output dir: %s", outputDir)

	downloadUrlPrefix = filepath.Join(downloadUrlPrefix, path)

	if !publicDeal && (minerFid == nil || len(*minerFid) == 0) {
		logs.GetLogger().Error("Please provide -miner for non public deal.")
		return nil, false
	}

	carFiles := ReadCarFilesFromJsonFile(*inputDir, JSON_FILE_NAME_BY_UPLOAD)
	if carFiles == nil {
		logs.GetLogger().Error("Failed to read car files from : ", *inputDir)
		return nil, false
	}

	task := model.Task{
		TaskName:       *taskName,
		CuratedDataset: *dataset,
		Description:    *description,
		//IsVerified:     verifiedDeal,
		FastRetrieval: &fastRetrieval,
		MaxPrice:      &maxPrice,
		BidMode:       &bidMode,
		ExpireDays:    &expireDays,
		MinerFid:      minerFid,
		Uuid:          uuid.NewString(),
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
			return nil, false
		}
	}

	jsonFileName := *taskName + JSON_FILE_NAME_BY_TASK_SUFFIX
	csvFileName := *taskName + CSV_FILE_NAME_BY_TASK_SUFFIX
	WriteCarFilesToFiles(carFiles, *outputDir, jsonFileName, csvFileName)
	SendTask2Swan(task, carFiles, *outputDir)
	return &jsonFileName, true
}

func SendTask2Swan(task model.Task, carFiles []*model.FileDesc, outDir string) bool {
	csvFilename := task.TaskName + "_task.csv"
	csvFilePath, err := CreateCsv4TaskDeal(carFiles, task.MinerFid, outDir, csvFilename)
	if err != nil {
		logs.GetLogger().Error("Failed to generate csv for task.")
		return false
	}

	if config.GetConfig().Sender.OfflineMode {
		logs.GetLogger().Info("Working in Offline Mode. You need to manually send out task on filwan.com.")
		return true
	}

	logs.GetLogger().Info("Working in Online Mode. A swan task will be created on the filwan.com after process done. ")

	swanClient := client.SwanGetClient()
	response := swanClient.SwanCreateTask(task, csvFilePath)
	logs.GetLogger().Info(response)

	return true
}
