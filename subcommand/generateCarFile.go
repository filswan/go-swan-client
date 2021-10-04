package subcommand

import (
	"go-swan-client/common/utils"
	"go-swan-client/config"
	"go-swan-client/logs"
	"io/ioutil"
	"strconv"

	"github.com/google/uuid"
)

func GenerateCarFiles(inputDir, outputDir *string) {
	if inputDir == nil || len(*inputDir) == 0 {
		logs.GetLogger().Error("Please provide input dir.")
		return
	}

	if !utils.IsFileExistsFullPath(*inputDir) {
		logs.GetLogger().Error("Input dir: ", inputDir, " not exists.")
		return
	}

	if outputDir == nil || len(*outputDir) == 0 {
		outDir := utils.GetDir(config.GetConfig().Sender.OutputDir, uuid.NewString())
		outputDir = &outDir
		logs.GetLogger().Info("output-dir is not provided, use default:", outputDir)
	}

	err := utils.CreateDir(*outputDir)
	if err != nil {
		logs.GetLogger().Error("Failed to create output dir:", outputDir)
		return
	}

	carFiles := []*FileDesc{}

	srcFiles, err := ioutil.ReadDir(*inputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}

	for _, f := range srcFiles {
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

		carFile.PieceCid = *pieceCid

		dataCid := utils.LotusImportCarFile(carFile.CarFilePath)
		if dataCid == nil {
			logs.GetLogger().Error("Failed to import car file.")
			return
		}

		carFile.DataCid = *dataCid

		carFile.CarFileSize = strconv.FormatInt(utils.GetFileSize(carFile.CarFilePath), 10)

		carFiles = append(carFiles, &carFile)
	}

	err = GenerateCsvFile(carFiles, *outputDir, "car.csv")
	if err != nil {
		logs.GetLogger().Error("Failed to create car file.")
		return
	}

	err = GenerateJsonFile(carFiles, *outputDir)
	if err != nil {
		logs.GetLogger().Error("Failed to create json file.")
		return
	}
}
