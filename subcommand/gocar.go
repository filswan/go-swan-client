package subcommand

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/DoraNebula/go-swan-client/common/constants"
	"github.com/DoraNebula/go-swan-client/config"
	"github.com/DoraNebula/go-swan-client/logs"
	"github.com/DoraNebula/go-swan-client/model"

	"github.com/DoraNebula/go-swan-client/common/utils"

	"github.com/DoraNebula/go-swan-client/common/client"

	"github.com/codingsince1985/checksum"
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

	srcFiles, err := ioutil.ReadDir(inputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	carDir := *outputDir
	for _, srcFile := range srcFiles {
		parentPath := filepath.Join(inputDir, srcFile.Name())
		targetPath := parentPath
		graphName := srcFile.Name()
		parallel := 4

		Emptyctx := context.Background()
		cb := graphsplit.CommPCallback(carDir)
		err = graphsplit.Chunk(Emptyctx, sliceSize, parentPath, targetPath, carDir, graphName, parallel, cb)
		if err != nil {
			logs.GetLogger().Error(err)
		}
	}
	logs.GetLogger().Info("car files generated")
	carFiles, err := CreateCarFilesDesc2Files(inputDir, carDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	return outputDir, carFiles, nil
}

type ManifestDetail struct {
	Name string
	Hash string
	Size int
	Link []ManifestDetailLinkItem
}

type ManifestDetailLinkItem struct {
	Name string
	Hash string
	Size int
}

func CreateCarFilesDesc2Files(srcFileDir, carFileDir string) ([]*model.FileDesc, error) {
	manifestFilename := "manifest.csv"
	lines, err := utils.ReadAllLines(carFileDir, manifestFilename)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	carFiles := []*model.FileDesc{}

	for i, line := range lines {
		if i == 0 {
			continue
		}

		fields := strings.Split(line, ",")
		if len(fields) < 5 {
			err := fmt.Errorf("not enough fields in %s", manifestFilename)
			logs.GetLogger().Error(err)
			return nil, err
		}

		carFile := model.FileDesc{}
		carFile.DataCid = fields[0]
		carFile.CarFileName = carFile.DataCid + ".car"
		carFile.CarFilePath = filepath.Join(carFileDir, carFile.CarFileName)
		carFile.PieceCid = fields[2]
		carFile.CarFileSize = utils.GetInt64FromStr(fields[3])

		carFileDetail := fields[4]
		for i := 5; i < len(fields); i++ {
			carFileDetail = carFileDetail + "," + fields[i]
		}
		logs.GetLogger().Info("carFileDetail:", carFileDetail)
		manifestDetail := ManifestDetail{}
		err = json.Unmarshal([]byte(carFileDetail), &manifestDetail)
		if err != nil {
			logs.GetLogger().Error("Failed to parse: ", carFileDetail)
			return nil, err
		}

		carFile.SourceFileName = manifestDetail.Link[0].Name
		carFile.SourceFilePath = filepath.Join(srcFileDir, carFile.SourceFileName)
		carFile.SourceFileSize = utils.GetFileSize(carFile.SourceFilePath)

		srcFileMd5, err := checksum.MD5sum(carFile.SourceFilePath)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}
		carFile.SourceFileMd5 = srcFileMd5

		carFileMd5, err := checksum.MD5sum(carFile.CarFilePath)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}
		carFile.CarFileMd5 = carFileMd5

		carFiles = append(carFiles, &carFile)
	}

	err = WriteCarFilesToFiles(carFiles, carFileDir, constants.JSON_FILE_NAME_BY_GOCAR, constants.CSV_FILE_NAME_BY_GOCAR)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	logs.GetLogger().Info("Car files output dir: ", carFileDir)
	logs.GetLogger().Info("Please upload car files to web server or ipfs server.")

	return carFiles, nil
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
