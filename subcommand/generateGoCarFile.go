package subcommand

import (
	"go-swan-client/common/utils"
	"go-swan-client/config"
	"go-swan-client/logs"
	"io/ioutil"

	"github.com/google/uuid"
)

func GenerateGoCarFiles(inputDir, outputDir *string) bool {
	if outputDir == nil {
		outDir := utils.GetDir(config.GetConfig().Sender.OutputDir, uuid.NewString())
		outputDir = &outDir
	}

	err := utils.CreateDir(*outputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return false
	}

	carFiles := []*FileDesc{}

	srcFiles, err := ioutil.ReadDir(*inputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return false
	}

	for _, srcFile := range srcFiles {
		carFile := FileDesc{}
		carFile.SourceFileName = srcFile.Name()
		carFile.SourceFilePath = utils.GetDir(*inputDir, carFile.SourceFileName)
		carFile.SourceFileSize = srcFile.Size()

		carFiles = append(carFiles, &carFile)
	}

	result := GenerateGoCar(carFiles, *outputDir)

	return result
}

func GenerateGoCar(carFiles []*FileDesc, outputDir string) bool {
	for _, carFile := range carFiles {
		carFile.CarFileName = carFile.SourceFileName + ".car"
		carFile.CarFilePath = utils.GetDir(outputDir, carFile.CarFileName)

		dataCid, carFileName, isSucceed := utils.GraphSlit(outputDir, carFile.SourceFileName, carFile.CarFilePath)
		if !isSucceed {
			logs.GetLogger().Error("Failed to generate car file.")
			return false
		}
		carFile.DataCid = *dataCid
		carFile.CarFileName = *carFileName
	}

	return true
}
