package subcommand

import (
	"go-swan-client/common/utils"
	"go-swan-client/config"
	"go-swan-client/logs"
	"io/ioutil"

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
			outDir := utils.GetDir(config.GetConfig().Sender.OutputDir, uuid.NewString())
			outputDir = &outDir
		} else {
			*outputDir = utils.GetDir(config.GetConfig().Sender.OutputDir, uuid.NewString())
		}

		logs.GetLogger().Info("output-dir is not provided, use default:", outputDir)
	}

	err := utils.CreateDir(*outputDir)
	if err != nil {
		logs.GetLogger().Fatal("Failed to create output dir:", outputDir)
	}

	carFiles := []*FileDesc{}

	srcFiles, err := ioutil.ReadDir(*inputDir)
	if err != nil {
		logs.GetLogger().Fatal(err)
	}

	for _, srcFile := range srcFiles {
		carFile := FileDesc{}
		carFile.SourceFileName = srcFile.Name()
		carFile.SourceFilePath = utils.GetDir(*inputDir, carFile.SourceFileName)
		carFile.SourceFileSize = srcFile.Size()
		carFile.CarFileName = carFile.SourceFileName + ".car"
		carFile.CarFilePath = utils.GetDir(*outputDir, carFile.CarFileName)

		err := utils.LotusGenerateCar(carFile.SourceFilePath, carFile.CarFilePath)
		if err != nil {
			logs.GetLogger().Fatal("Failed to generate car file.")
		}

		pieceCid := utils.LotusGeneratePieceCid(carFile.CarFilePath)
		if pieceCid == nil {
			logs.GetLogger().Fatal("Failed to generate piece cid.")
		}

		carFile.PieceCid = *pieceCid

		dataCid := utils.LotusImportCarFile(carFile.CarFilePath)
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

	err = GenerateCsvFile(carFiles, *outputDir, "car.csv")
	if err != nil {
		logs.GetLogger().Fatal("Failed to create car file.")
	}

	err = GenerateJsonFile(carFiles, *outputDir)
	if err != nil {
		logs.GetLogger().Fatal("Failed to create json file.")
	}
}
