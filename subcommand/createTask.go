package subcommand

import (
	"encoding/csv"
	"go-swan-client/common/constants"
	"go-swan-client/common/utils"
	"go-swan-client/config"
	"go-swan-client/logs"
	"go-swan-client/models"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func CreateTask(taskName, inputDir, outputDir, minerFid, dataset, description *string) {
	if outputDir == nil {
		outDir := config.GetConfig().Sender.OutputDir
		outputDir = &outDir
	}
	publicDeal := config.GetConfig().Sender.PublicDeal
	verifiedDeal := config.GetConfig().Sender.VerifiedDeal
	//generateMd5 := config.GetConfig().Sender.GenerateMd5
	offlineMode := config.GetConfig().Sender.OfflineMode

	storageServerType := config.GetConfig().Main.StorageServerType
	host := config.GetConfig().WebServer.Host
	port := config.GetConfig().WebServer.Port
	path := config.GetConfig().WebServer.Path

	downloadUrlPrefix := strings.TrimRight(host, "/") + ":" + strconv.Itoa(port)
	taskUuid := uuid.NewString()

	path = strings.TrimRight(path, "/")
	//finalCsvPath := ""

	logs.GetLogger().Info("Swan Client Settings: Public Task: ", publicDeal, ",  Verified Deals: ", verifiedDeal, ",  Connected to Swan: ", !offlineMode, ", CSV/car File output dir: %s", outputDir)

	if path != "" {
		downloadUrlPrefix = utils.GetDir(downloadUrlPrefix, path)
	}

	if !publicDeal && minerFid == nil {
		logs.GetLogger().Error("Please provide --miner for non public deal.")
		return
	}

	carFiles := ReadCarFilesFromJsonFile(*inputDir)
	if carFiles == nil {
		logs.GetLogger().Error("Failed to read car files from : ", inputDir)
		return
	}

	for _, carFile := range carFiles {
		carFile.Uuid = taskUuid
	}

	if storageServerType == constants.STORAGE_SERVER_TYPE_WEB_SERVER {
		for _, carFile := range carFiles {
			carFile.CarFileUrl = utils.GetDir(downloadUrlPrefix, carFile.CarFileName)
		}
	}

	err := utils.CreateDir(*outputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}

	if offlineMode {
		logs.GetLogger().Info("Working in Offline Mode. You need to manually send out task on filwan.com.")
	} else {
		logs.GetLogger().Info("Working in Online Mode. A swan task will be created on the filwan.com after process done. ")
	}

	task := models.Task{}
	task.TaskName = *taskName
	task.CuratedDataset = *dataset
	task.Description = *description
	if publicDeal {
		task.IsPublic = 1
	} else {
		task.IsPublic = 0
	}
	task.IsVerified = verifiedDeal
}

func GenerateMetadataCsv(task models.Task, carFiles []*FileDesc, outDir string) error {
	csvFileName := task.TaskName + "-metadata.csv"
	csvFilePath := utils.GetDir(outDir, csvFileName)

	err := GenerateCsvFile(carFiles, outDir, csvFileName)
	if err != nil {
		logs.GetLogger().Error("Failed to generate metadata csv file.")
		return err
	}

	logs.GetLogger().Info("Metadata CSV Generated: ", csvFilePath)

	return nil
}

func SendTask2Swan(task models.Task, carFiles []*FileDesc, outDir string) error {
	csvFileName := task.TaskName + ".csv"
	csvFilePath := utils.GetDir(outDir, csvFileName)

	logs.GetLogger().Info("Swan task CSV Generated: ", csvFilePath)

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
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.Write(headers)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	for _, carFile := range carFiles {
		columns := []string{
			carFile.Uuid,
			carFile.MinerId,
			carFile.DealCid,
			"payload_cid",
			carFile.CarFileUrl,
			carFile.CarFileMd5,
			carFile.StartEpoch,
		}

		err = writer.Write(columns)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
	}

	swanClient := utils.SwanGetClient()

	ioutil.ReadFile("")
	response := swanClient.SwanCreateTask(task, csvFilePath)
	logs.GetLogger().Info(response)

	return nil
}
