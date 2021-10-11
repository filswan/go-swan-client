package subcommand

import (
	"go-swan-client/common/client"
	"go-swan-client/common/constants"
	"go-swan-client/config"
	"go-swan-client/logs"
	"go-swan-client/model"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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
		IsPublic:       publicDeal,
		IsVerified:     verifiedDeal,
		MinerId:        minerFid,
		Uuid:           uuid.NewString(),
	}

	for _, carFile := range carFiles {
		carFile.Uuid = task.Uuid

		if storageServerType == constants.STORAGE_SERVER_TYPE_WEB_SERVER {
			carFile.CarFileUrl = filepath.Join(downloadUrlPrefix, carFile.CarFileName)
		}
	}

	if !publicDeal {
		result := SendDeals2Miner(nil, *taskName, *minerFid, *outputDir, carFiles)
		if !result {
			return nil, result
		}
	}

	jsonFileName := *taskName + JSON_FILE_NAME_BY_TASK_SUFFIX
	csvFileName := *taskName + CSV_FILE_NAME_BY_TASK_SUFFIX
	WriteCarFilesToFiles(carFiles, *outputDir, jsonFileName, csvFileName)
	SendTask2Swan(task, carFiles, *outputDir)
	return &jsonFileName, true
}

func SendTask2Swan(task model.Task, carFiles []*model.FileDesc, outDir string) bool {
	csvFilePath, err := CreateCsv4TaskDeal(task.TaskName, carFiles, task.MinerId, outDir)
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
