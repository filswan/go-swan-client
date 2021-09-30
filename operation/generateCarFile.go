package operation

import (
	"encoding/csv"
	"go-swan-client/common/utils"
	"go-swan-client/config"
	"go-swan-client/logs"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/google/uuid"
)

type FileDesc struct {
	CarFileName    string
	CarFilePath    string
	pieceCid       string
	DataCid        string
	CarFileSize    string
	CarFileMd5     string
	SourceFileName string
	SourceFilePath string
	SourceFileSize string
	SourceFileMd5  string
	CarFileAddress string
}

func GenerateCarFiles(inputDir *string, outputDir *string) {
	if inputDir == nil {
		logs.GetLogger().Error("Please provide input dir.")
		return
	}

	if !utils.IsFileExistsFullPath(*inputDir) {
		logs.GetLogger().Error("Input dir: ", *inputDir, " not exists.")
		return
	}

	if outputDir == nil {
		outDir := utils.GetDir(config.GetConfig().Sender.OutputDir, uuid.NewString())
		outputDir = &outDir
	}

	err := utils.CreateDir(*outputDir)
	if err != nil {
		logs.GetLogger().Error("Failed to create output dir:", outputDir)
		return
	}

	carFiles := []*FileDesc{}

	files, err := ioutil.ReadDir(*inputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}

	for _, f := range files {
		carFile := FileDesc{}
		carFile.SourceFileName = f.Name()
		carFile.SourceFilePath = utils.GetDir(*inputDir, carFile.SourceFileName)
		carFile.SourceFileSize = strconv.FormatInt(utils.GetFileSize(carFile.SourceFilePath), 10)
		carFile.CarFileName = carFile.SourceFileName + ".car"
		carFile.CarFilePath = utils.GetDir(*outputDir, carFile.CarFileName)

		isCarGenerated := utils.LotusGenerateCar(carFile.SourceFilePath, carFile.CarFilePath)
		if !isCarGenerated {
			logs.GetLogger().Error("Failed to generate car file.")
			return
		}

		pieceCid, pieceSize := utils.LotusGeneratePieceCid(carFile.CarFilePath)
		if pieceCid == nil || pieceSize == nil {
			logs.GetLogger().Error("Failed to generate piece cid.")
			return
		}

		carFile.pieceCid = *pieceCid

		dataCid := utils.LotusImportCarFile(carFile.CarFilePath)
		if dataCid == nil {
			logs.GetLogger().Error("Failed to import car file.")
			return
		}

		carFile.DataCid = *dataCid

		carFile.CarFileSize = strconv.FormatInt(utils.GetFileSize(carFile.CarFilePath), 10)

		carFiles = append(carFiles, &carFile)
	}

	err = GenerateSummaryFile(carFiles, *outputDir)
	if err != nil {
		logs.GetLogger().Error("Failed to create car files.")
	}
}

func GenerateSummaryFile(carFiles []*FileDesc, outputDir string) error {
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
		var columns []string
		columns = append(columns, carFile.CarFileName)
		columns = append(columns, carFile.CarFilePath)
		columns = append(columns, carFile.pieceCid)
		columns = append(columns, carFile.DataCid)
		columns = append(columns, carFile.CarFileSize)
		columns = append(columns, carFile.CarFileMd5)
		columns = append(columns, carFile.SourceFileName)
		columns = append(columns, carFile.SourceFilePath)
		columns = append(columns, carFile.CarFileSize)
		columns = append(columns, carFile.SourceFileMd5)
		columns = append(columns, "")

		err = writer.Write(columns)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
	}

	logs.GetLogger().Info("Car files output dir: ", outputDir)
	logs.GetLogger().Info("Please upload car files to web server or ipfs server.")

	return nil
}
