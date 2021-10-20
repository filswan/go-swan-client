package subcommand

import (
	"context"
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

func GoCar() {
	Emptyctx := context.Background()
	var cb graphsplit.GraphBuildCallback

	sliceSize := int64(10)
	carDir := "/home/peware/go-swan-client/carFiles"
	parentPath := "/home/peware/go-swan-client/srcFiles"
	targetPath := "/home/peware/go-swan-client/srcFiles"
	graphName := "test"
	parallel := 4

	cb = graphsplit.CommPCallback(carDir)
	err := graphsplit.Chunk(Emptyctx, sliceSize, parentPath, targetPath, carDir, graphName, parallel, cb)
	if err != nil {
		logs.GetLogger().Error(err)
	}
	logs.GetLogger().Info("car files generated")
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
