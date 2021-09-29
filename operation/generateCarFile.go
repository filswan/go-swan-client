package operation

import (
	"encoding/csv"
	"go-swan-client/common/utils"
	"go-swan-client/config"
	"go-swan-client/logs"
	"go-swan-client/models"
	"io/ioutil"
	"os"
	"strconv"
)

func GenerateCarFiles(inputDir string, outDir *string) {
	var outputDir string
	if outDir == nil {
		outputDir = config.GetConfig().Sender.OutputDir //+ '/' + str(uuid.uuid4())
	} else {
		outputDir = *outDir
	}

	err := utils.CreateDir(outputDir)
	if err != nil {
		logs.GetLogger().Error("Failed to create output dir:", outputDir)
		return
	}

	offlineDeals := []models.OfflineDeal{}

	files, err := ioutil.ReadDir(inputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}

	//generateMd5 := config.GetConfig().Sender.GenerateMd5
	for _, f := range files {
		offlineDeal := models.OfflineDeal{}
		offlineDeal.FileName = f.Name()
		offlineDeal.FilePath = utils.GetDir(inputDir, offlineDeal.FileName)
		fileSize := strconv.FormatInt(utils.GetFileSize(offlineDeal.FilePath), 10)
		offlineDeal.FileSize = &fileSize
		offlineDeals = append(offlineDeals, offlineDeal)
	}

	GenerateCar(offlineDeals, *outDir)
}

func GenerateCar(offlineDeals []models.OfflineDeal, outputDir string) {
	csvPath := utils.GetDir(outputDir, "car.csv")

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

	file, err := os.Create(csvPath)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.Write(headers)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}

	for _, offlineDeal := range offlineDeals {
		offlineDeal.CarFileName = offlineDeal.FileName + ".car"
		offlineDeal.CarFilePath = utils.GetDir(outputDir, offlineDeal.CarFileName)
		utils.RemoveFile(outputDir, offlineDeal.CarFileName)

		carFileSize := utils.GetFileSize(offlineDeal.FilePath)
		carMd5 := "checksum(car_file_path)"
		var columns []string
		columns = append(columns, offlineDeal.CarFileName)
		columns = append(columns, offlineDeal.CarFilePath)
		columns = append(columns, *offlineDeal.PieceCid)
		columns = append(columns, offlineDeal.DealCid)
		columns = append(columns, strconv.FormatInt(carFileSize, 10))
		columns = append(columns, carMd5)
		columns = append(columns, offlineDeal.FileName)
		columns = append(columns, offlineDeal.FilePath)
		columns = append(columns, *offlineDeal.FileSize)
		columns = append(columns, "source file md5")
		columns = append(columns, "")
	}

	logs.GetLogger().Info("Car files output dir: ", outputDir)
	logs.GetLogger().Info("Please upload car files to web server or ipfs server.")
}
