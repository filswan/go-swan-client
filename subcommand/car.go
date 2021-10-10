package subcommand

import (
	"go-swan-client/common/client"
	"go-swan-client/common/utils"
	"go-swan-client/config"
	"go-swan-client/logs"
	"go-swan-client/model"
	"io/ioutil"
	"os"

	"github.com/codingsince1985/checksum"
	"github.com/google/uuid"
)

func GenerateCarFiles(inputDir, outputDir *string) {
	if inputDir == nil || len(*inputDir) == 0 {
		logs.GetLogger().Fatal("Please provide input dir.")
	}

	if !utils.IsFileExistsFullPath(*inputDir) {
		logs.GetLogger().Fatal("Input dir: ", inputDir, " not exists.")
	}

	if outputDir == nil || len(*outputDir) == 0 {
		if outputDir == nil {
			outDir := utils.GetPath(config.GetConfig().Sender.OutputDir, uuid.NewString())
			outputDir = &outDir
		} else {
			*outputDir = utils.GetPath(config.GetConfig().Sender.OutputDir, uuid.NewString())
		}

		logs.GetLogger().Info("output-dir is not provided, use default:", outputDir)
	}

	err := os.MkdirAll(*outputDir, os.ModePerm)
	if err != nil {
		logs.GetLogger().Fatal("Failed to create output dir:", outputDir)
	}

	carFiles := []*model.FileDesc{}

	srcFiles, err := ioutil.ReadDir(*inputDir)
	if err != nil {
		logs.GetLogger().Fatal(err)
	}

	for _, srcFile := range srcFiles {
		carFile := model.FileDesc{}
		carFile.SourceFileName = srcFile.Name()
		carFile.SourceFilePath = utils.GetPath(*inputDir, carFile.SourceFileName)
		carFile.SourceFileSize = srcFile.Size()
		carFile.CarFileName = carFile.SourceFileName + ".car"
		carFile.CarFilePath = utils.GetPath(*outputDir, carFile.CarFileName)

		err := client.LotusGenerateCar(carFile.SourceFilePath, carFile.CarFilePath)
		if err != nil {
			logs.GetLogger().Fatal("Failed to generate car file.")
		}

		pieceCid := client.LotusGeneratePieceCid(carFile.CarFilePath)
		if pieceCid == nil {
			logs.GetLogger().Fatal("Failed to generate piece cid.")
		}

		carFile.PieceCid = *pieceCid

		dataCid := client.LotusImportCarFile(carFile.CarFilePath)
		if dataCid == nil {
			logs.GetLogger().Fatal("Failed to import car file.")
		}

		carFile.DataCid = *dataCid

		carFile.CarFileSize = utils.GetFileSize(carFile.CarFilePath)

		if config.GetConfig().Sender.GenerateMd5 {
			carFileMd5, err := checksum.MD5sum(carFile.CarFilePath)
			if err != nil {
				logs.GetLogger().Error(err)
				logs.GetLogger().Fatal("Failed to generate md5 for car file:", carFile.CarFilePath)
			}
			logs.GetLogger().Info("carFileMd5:", carFileMd5)
			carFile.CarFileMd5 = carFileMd5
		}

		carFiles = append(carFiles, &carFile)
	}

	generateCsvFile(carFiles, *outputDir, "car.csv")
	generateJsonFile(carFiles, *outputDir)

	logs.GetLogger().Info("Car files output dir: ", outputDir)
	logs.GetLogger().Info("Please upload car files to web server or ipfs server.")
}
