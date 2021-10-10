package subcommand

import (
	"go-swan-client/common/client"
	"go-swan-client/common/utils"
	"go-swan-client/config"
	"go-swan-client/logs"
	"go-swan-client/model"
	"io/ioutil"
	"os"

	"github.com/google/uuid"
)

func GenerateGoCarFiles(inputDir, outputDir *string) bool {
	if outputDir == nil {
		outDir := utils.GetPath(config.GetConfig().Sender.OutputDir, uuid.NewString())
		outputDir = &outDir
	}

	err := os.MkdirAll(*outputDir, os.ModePerm)
	if err != nil {
		logs.GetLogger().Error(err)
		return false
	}

	carFiles := []*model.FileDesc{}

	srcFiles, err := ioutil.ReadDir(*inputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return false
	}

	for _, srcFile := range srcFiles {
		carFile := model.FileDesc{}
		carFile.SourceFileName = srcFile.Name()
		carFile.SourceFilePath = utils.GetPath(*inputDir, carFile.SourceFileName)
		carFile.SourceFileSize = srcFile.Size()

		carFiles = append(carFiles, &carFile)
	}

	result := GenerateGoCar(carFiles, *outputDir)

	return result
}

func GenerateGoCar(carFiles []*model.FileDesc, outputDir string) bool {
	for _, carFile := range carFiles {
		carFile.CarFileName = carFile.SourceFileName + ".car"
		carFile.CarFilePath = utils.GetPath(outputDir, carFile.CarFileName)

		dataCid, carFileName, isSucceed := client.GraphSlit(outputDir, carFile.SourceFileName, carFile.CarFilePath)
		if !isSucceed {
			logs.GetLogger().Error("Failed to generate car file.")
			return false
		}
		carFile.DataCid = *dataCid
		carFile.CarFileName = *carFileName
	}

	return true
}
