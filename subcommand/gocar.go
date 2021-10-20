package subcommand

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"go-swan-client/common/client"
	"go-swan-client/config"
	"go-swan-client/logs"
	"go-swan-client/model"

	"github.com/filedrive-team/go-graphsplit"
	"github.com/google/uuid"
)

func CreateGoCarFiles(inputDir string, outputDir *string) (*string, []*model.FileDesc, error) {
	err := CheckInputDir(inputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	outputDir, err = CreateOutputDir(outputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	sliceSize := config.GetConfig().Sender.GocarFileSizeLimit
	if sliceSize <= 0 {
		err := fmt.Errorf("gocar file size limit is too smal")
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	carFiles := []*model.FileDesc{}

	srcFiles, err := ioutil.ReadDir(inputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	for _, srcFile := range srcFiles {
		carFile := model.FileDesc{}
		carFile.SourceFileName = srcFile.Name()
		carFile.SourceFilePath = filepath.Join(inputDir, carFile.SourceFileName)
		carFile.SourceFileSize = srcFile.Size()

		carFile.CarFileName = carFile.SourceFileName + ".car"
		carFile.CarFilePath = filepath.Join(inputDir, carFile.CarFileName)

		carFiles = append(carFiles, &carFile)

		carDir := *outputDir
		parentPath := carFile.SourceFilePath
		targetPath := carFile.CarFilePath
		graphName := "test"
		parallel := 4

		Emptyctx := context.Background()
		cb := graphsplit.CommPCallback(carDir)
		err = graphsplit.Chunk(Emptyctx, sliceSize, parentPath, targetPath, carDir, graphName, parallel, cb)
		if err != nil {
			logs.GetLogger().Error(err)
		}
	}
	logs.GetLogger().Info("car files generated")

	return outputDir, nil, nil
}

func GenerateGoCarFiles(inputDir, outputDir *string) bool {
	if outputDir == nil {
		outDir := filepath.Join(config.GetConfig().Sender.OutputDir, uuid.NewString())
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
		carFile.SourceFilePath = filepath.Join(*inputDir, carFile.SourceFileName)
		carFile.SourceFileSize = srcFile.Size()

		carFiles = append(carFiles, &carFile)
	}

	result := GenerateGoCar(carFiles, *outputDir)

	return result
}

func GenerateGoCar(carFiles []*model.FileDesc, outputDir string) bool {
	for _, carFile := range carFiles {
		carFile.CarFileName = carFile.SourceFileName + ".car"
		carFile.CarFilePath = filepath.Join(outputDir, carFile.CarFileName)

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
