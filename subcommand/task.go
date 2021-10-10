package subcommand

import (
	"encoding/csv"
	"go-swan-client/common/client"
	"go-swan-client/common/constants"
	"go-swan-client/common/utils"
	"go-swan-client/config"
	"go-swan-client/logs"
	"go-swan-client/models"
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

	carFiles := readCarFilesFromJsonFile(*inputDir)
	if carFiles == nil {
		logs.GetLogger().Error("Failed to read car files from : ", *inputDir)
		return false
	}

	if storageServerType == constants.STORAGE_SERVER_TYPE_WEB_SERVER {
		for _, carFile := range carFiles {
			carFile.CarFileUrl = utils.GetPath(downloadUrlPrefix, carFile.CarFileName)
		}
	}

	if !publicDeal {
		//sendDeals(outputDir,task)
	}

	err := os.MkdirAll(*outputDir, os.ModePerm)
	if err != nil {
		logs.GetLogger().Error(err)
		logs.GetLogger().Error("Failed to create output dir:", *outputDir)
		return false
	}

	task := models.Task{}
	task.TaskName = *taskName
	task.CuratedDataset = *dataset
	task.Description = *description
	task.IsPublic = publicDeal
	task.IsVerified = verifiedDeal
	task.MinerId = minerFid

	taskUuid := uuid.NewString()
	for _, carFile := range carFiles {
		carFile.Uuid = taskUuid
	}

	GenerateMetadataCsv(task, carFiles, *outputDir)
	SendTask2Swan(task, carFiles, *outputDir)
	return true
}

func GenerateMetadataCsv(task models.Task, carFiles []*FileDesc, outDir string) error {
	csvFilePath := utils.GetPath(outDir, task.TaskName+"-metadata.csv")
	var headers []string
	headers = append(headers, "uuid")
	headers = append(headers, "source_file_name")
	headers = append(headers, "source_file_path")
	headers = append(headers, "source_file_md5")
	headers = append(headers, "source_file_url")
	headers = append(headers, "source_file_size")
	headers = append(headers, "car_file_name")
	headers = append(headers, "car_file_path")
	headers = append(headers, "car_file_md5")
	headers = append(headers, "car_file_url")
	headers = append(headers, "car_file_size")
	headers = append(headers, "deal_cid")
	headers = append(headers, "data_cid")
	headers = append(headers, "piece_cid")
	headers = append(headers, "miner_id")
	headers = append(headers, "start_epoch")

	file, err := os.Create(csvFilePath)
	if err != nil {
		logs.GetLogger().Fatal(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.Write(headers)
	if err != nil {
		logs.GetLogger().Fatal(err)
	}

	for _, carFile := range carFiles {
		var columns []string
		columns = append(columns, carFile.Uuid)
		columns = append(columns, carFile.SourceFileName)
		columns = append(columns, carFile.SourceFilePath)
		columns = append(columns, strconv.FormatBool(carFile.SourceFileMd5))
		columns = append(columns, carFile.SourceFileUrl)
		columns = append(columns, strconv.FormatInt(carFile.SourceFileSize, 10))
		columns = append(columns, carFile.CarFileName)
		columns = append(columns, carFile.CarFilePath)
		columns = append(columns, carFile.CarFileMd5)
		columns = append(columns, carFile.CarFileUrl)
		columns = append(columns, strconv.FormatInt(carFile.CarFileSize, 10))
		columns = append(columns, carFile.DealCid)
		columns = append(columns, carFile.DataCid)
		columns = append(columns, carFile.PieceCid)
		if task.MinerId != nil {
			columns = append(columns, *task.MinerId)
		} else {
			columns = append(columns, "")
		}
		columns = append(columns, carFile.StartEpoch)

		err = writer.Write(columns)
		if err != nil {
			logs.GetLogger().Fatal(err)
		}
	}

	logs.GetLogger().Info("Metadata CSV Generated: ", csvFilePath)

	return nil
}

func SendTask2Swan(task models.Task, carFiles []*FileDesc, outDir string) bool {
	csvFileName := task.TaskName + ".csv"
	csvFilePath := filepath.Join(outDir, csvFileName)

	headers := []string{
		"uuid",
		"miner_id",
		"deal_cid",
		"payload_cid",
		"file_source_url",
		"md5",
		"start_epoch",
	}

	file, err := os.Create(csvFilePath)
	if err != nil {
		logs.GetLogger().Error(err)
		logs.GetLogger().Error("Failed to genereate:", csvFilePath)
		return false
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

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
