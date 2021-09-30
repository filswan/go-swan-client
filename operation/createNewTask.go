package operation

import (
	"encoding/csv"
	"go-swan-client/common/utils"
	"go-swan-client/config"
	"go-swan-client/logs"
	"go-swan-client/models"
	"os"
	"strconv"
	"strings"

	"io/ioutil"

	"github.com/google/uuid"
)

type Deal struct {
	carFileName    string
	carfilePath    string
	pieceCid       string
	dataCid        string
	carFileSize    string
	carFileMd5     string
	sourceFileName string
	sourceFilePath string
	sourceFileSize string
	sourceFileMd5  string
	carFileUrl     string
	uuid           string
}

type Csv struct {
	uuid          string
	minerId       string
	dealCid       string
	payloadCid    string
	fileSourceUrl string
	md5           string
	startEpoch    string
}

func CreateNewTask(inputDir, outDir, configPath, taskName, curatedDataset, description string, minerId *int) {
	outputDir := outDir
	if outDir == "" {
		outputDir = config.GetConfig().Sender.OutputDir
	}
	publicDeal := config.GetConfig().Sender.PublicDeal
	verifiedDeal := config.GetConfig().Sender.VerifiedDeal
	generateMd5 := config.GetConfig().Sender.GenerateMd5
	offlineMode := config.GetConfig().Sender.OfflineMode

	//apiUrl := config.GetConfig().Main.SwanApiUrl
	//apiKey := config.GetConfig().Main.SwanApiKey
	//accessToken := config.GetConfig().Main.SwanAccessToken

	storageServerType := config.GetConfig().Main.StorageServerType
	host := config.GetConfig().WebServer.Host
	port := config.GetConfig().WebServer.Port
	path := config.GetConfig().WebServer.Path

	downloadUrlPrefix := strings.TrimRight(host, "/") + ":" + strconv.Itoa(port)
	taskUuid := uuid.New().String()
	//finalCsvPath := ""

	path = strings.TrimRight(path, "/")

	logs.GetLogger().Info("Swan Client Settings: Public Task: ", publicDeal, ",  Verified Deals: ", verifiedDeal, ",  Connected to Swan: ", !offlineMode, ", CSV/car File output dir: %s", outputDir)

	if path != "" {
		downloadUrlPrefix = utils.GetDir(downloadUrlPrefix, path)
	}

	if !publicDeal && minerId == nil {
		logs.GetLogger().Error("Please provide --miner for non public deal.")
		return
	}

	files, err := ioutil.ReadDir(inputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}

	utils.CreateDir(outputDir)

	offlineDeals := []*models.OfflineDeal{}

	for _, f := range files {
		offlineDeal := models.OfflineDeal{}
		offlineDeal.SourceFileName = f.Name()
		offlineDeal.SourceFilePath = inputDir
		offlineDeal.SourceFileSize = int(utils.GetFileSize2(offlineDeal.SourceFilePath, offlineDeal.SourceFileName))
		offlineDeal.CarFileMd5 = generateMd5
		offlineDeals = append(offlineDeals, &offlineDeal)
	}

	carFiles, err := utils.ReadAllLines(inputDir, "car.csv")
	if err != nil {
		logs.GetLogger().Info(err)
		return
	}

	deals := []*Deal{}

	for i := 1; i < len(carFiles); i++ {
		fileInfo := carFiles[i]
		fields := strings.Split(fileInfo, ",")
		deal := &Deal{
			carFileName:    fields[0],
			carfilePath:    fields[1],
			pieceCid:       fields[2],
			dataCid:        fields[3],
			carFileSize:    fields[4],
			carFileMd5:     fields[5],
			sourceFileName: fields[6],
			sourceFilePath: fields[7],
			sourceFileSize: fields[8],
			sourceFileMd5:  fields[9],
			carFileUrl:     fields[10],
		}

		if storageServerType == "web server" {
			deal.carFileUrl = utils.GetDir(downloadUrlPrefix, deal.carFileName)
		}

		if !publicDeal {
			//final_csv_path = send_deals(config_path, miner_id, task_name, deal_list=deal_list, task_uuid=task_uuid, out_dir=output_dir)
		}

		if offlineMode {
			logs.GetLogger().Info("Working in Offline Mode. You need to manually send out task on filwan.com. ")
		} else {
			//client = SwanClient(api_url, api_key, access_token)
			logs.GetLogger().Info("Working in Online Mode. A swan task will be created on the filwan.com after process done. ")
		}

		deals = append(deals, deal)
	}

	task := models.Task{
		TaskName:       taskName,
		CuratedDataset: curatedDataset,
		Description:    description,
	}

	if publicDeal {
		task.IsPublic = 1
	} else {
		task.IsPublic = 0
	}

	if minerId != nil {
		task.MinerId = minerId
	}

	client := utils.GetSwanClient()

	GenerateMetadataCsv(deals, task, outDir, taskUuid)
	GenerateCsvAndSend(task, deals, outDir, client)

}

func GenerateMetadataCsv(deals []*Deal, task models.Task, outDir string, uuid string) {
	csvFileName := task.TaskName + "-metadata.csv"
	csvFilePath := utils.GetDir(outDir, csvFileName)

	logs.GetLogger().Info("Metadata CSV Generated: ", csvFilePath)

	file, err := os.Create(csvFilePath)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	var headers []string
	headers = append(headers, "car_file_name")
	headers = append(headers, "car_file_path")
	headers = append(headers, "piece_cid")
	headers = append(headers, "data_cid")
	headers = append(headers, "car_file_size")
	headers = append(headers, "car_file_md5")
	headers = append(headers, "source_file_name")
	headers = append(headers, "source_file_path")
	headers = append(headers, "source_file_size")
	headers = append(headers, "source_file_md5")
	headers = append(headers, "car_file_url")

	err = writer.Write(headers)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}

	for _, deal := range deals {
		deal.uuid = uuid
		var columns []string
		//columns = append(columns, deal.CarFileName)
		//columns = append(columns, offlineDeal.CarFilePath)
		//columns = append(columns, *offlineDeal.PieceCid)
		//columns = append(columns, offlineDeal.DealCid)
		//columns = append(columns, strconv.FormatInt(carFileSize, 10))
		//columns = append(columns, carMd5)
		//columns = append(columns, offlineDeal.FileName)
		//columns = append(columns, offlineDeal.FilePath)
		//columns = append(columns, *offlineDeal.FileSize)
		//columns = append(columns, "source file md5")
		//columns = append(columns, "")
		err = writer.Write(columns)
	}
}

func GenerateCsvAndSend(task models.Task, deals []*Deal, outDir string, swanClient *utils.SwanClient) {
	csvFileName := task.TaskName + ".csv"
	csvFilePath := utils.GetDir(outDir, csvFileName)

	logs.GetLogger().Info("Swan task CSV Generated: ", csvFilePath)

	//fileInfos, err := utils.ReadAllLines(outDir, csvFileName)
	//if err != nil {
	//	logs.GetLogger().Info(err)
	//	return
	//}
	//for _, fileInfo := range fileInfos {
	//fields := strings.Split(fileInfo, ",")

	//csvData := Csv{
	//	uuid:          fields[0],
	//	minerId:       fields[1],
	//	dealCid:       fields[2],
	//	payloadCid:    fields[3],
	//	fileSourceUrl: fields[4],
	//	md5:           fields[5],
	//	startEpoch:    fields[0],
	//}
	//}

	swanClient.PostTask(task)
}
