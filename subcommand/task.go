package subcommand

import (
	"encoding/csv"
	"go-swan-client/common/client"
	"go-swan-client/common/constants"
	"go-swan-client/config"
	"go-swan-client/logs"
	"go-swan-client/model"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func CreateTask(taskName, inputDir, outputDir, minerFid, dataset, description *string) bool {
	if outputDir == nil {
		outDir := config.GetConfig().Sender.OutputDir
		outputDir = &outDir
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

	if !publicDeal && minerFid == nil {
		logs.GetLogger().Error("Please provide -miner for non public deal.")
		return false
	}

	carFiles := ReadCarFilesFromJsonFile(*inputDir, JSON_FILE_NAME_AFTER_UPLOAD)
	if carFiles == nil {
		logs.GetLogger().Error("Failed to read car files from : ", *inputDir)
		return false
	}

	err := os.MkdirAll(*outputDir, os.ModePerm)
	if err != nil {
		logs.GetLogger().Error(err)
		logs.GetLogger().Error("Failed to create output dir:", *outputDir)
		return false
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
		result := SendDeals2Miner(*minerFid, *outputDir, carFiles)
		if !result {
			return result
		}
	}

	GenerateMetadataCsv(task.MinerId, carFiles, *outputDir, task.TaskName+"-metadata.csv")
	WriteCarFilesToJsonFile(carFiles, *outputDir, JSON_FILE_NAME_AFTER_TASK)
	SendTask2Swan(task, carFiles, *outputDir)
	return true
}

func SendTask2Swan(task model.Task, carFiles []*model.FileDesc, outDir string) bool {
	csvFileName := task.TaskName + ".csv"
	csvFilePath := filepath.Join(outDir, csvFileName)

	file, err := os.Create(csvFilePath)
	if err != nil {
		logs.GetLogger().Error(err)
		logs.GetLogger().Error("Failed to genereate:", csvFilePath)
		return false
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{
		"uuid",
		"miner_id",
		"deal_cid",
		"payload_cid",
		"file_source_url",
		"md5",
		"start_epoch",
	}

	err = writer.Write(headers)
	if err != nil {
		logs.GetLogger().Error(err)
		logs.GetLogger().Error("Failed to genereate:", csvFilePath)
		return false
	}

	for _, carFile := range carFiles {
		var columns []string
		columns = append(columns, carFile.Uuid)
		if task.MinerId != nil {
			columns = append(columns, *task.MinerId)
		} else {
			columns = append(columns, "")
		}
		columns = append(columns, carFile.DealCid)
		columns = append(columns, carFile.DataCid)
		columns = append(columns, carFile.CarFileUrl)
		columns = append(columns, carFile.CarFileMd5)
		columns = append(columns, carFile.StartEpoch)

		err = writer.Write(columns)
		if err != nil {
			logs.GetLogger().Error(err)
			logs.GetLogger().Error("Failed to genereate:", csvFilePath)
			return false
		}
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
