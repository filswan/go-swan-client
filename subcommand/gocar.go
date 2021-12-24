package subcommand

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/filswan/go-swan-client/model"

	"github.com/filswan/go-swan-lib/logs"

	"github.com/filswan/go-swan-client/common/constants"
	"github.com/filswan/go-swan-lib/utils"

	"github.com/codingsince1985/checksum"
	"github.com/filedrive-team/go-graphsplit"
	"github.com/filswan/go-swan-lib/client/lotus"
	libmodel "github.com/filswan/go-swan-lib/model"
)

func CreateGoCarFilesByConfig(inputDir string, outputDir *string) ([]*libmodel.FileDesc, error) {
	confCar := model.GetConfCar(inputDir, outputDir)
	fileDescs, err := CreateGoCarFiles(confCar)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return fileDescs, nil
}

func CreateGoCarFiles(confCar *model.ConfCar) ([]*libmodel.FileDesc, error) {
	if confCar == nil {
		err := fmt.Errorf("parameter confCar is nil")
		logs.GetLogger().Error(err)
		return nil, err
	}

	err := CheckInputDir(confCar.InputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	err = CreateOutputDir(confCar.OutputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	sliceSize := confCar.GocarFileSizeLimit
	if sliceSize <= 0 {
		err := fmt.Errorf("gocar file size limit is too smal")
		logs.GetLogger().Error(err)
		return nil, err
	}

	srcFiles, err := ioutil.ReadDir(confCar.InputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	carDir := confCar.OutputDir
	for _, srcFile := range srcFiles {
		parentPath := filepath.Join(confCar.InputDir, srcFile.Name())
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
	carFiles, err := CreateCarFilesDescFromGoCarManifest(confCar, confCar.InputDir, carDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	logs.GetLogger().Info(len(carFiles), " car files have been created to directory:", carDir)
	logs.GetLogger().Info("Please upload car files to web server or ipfs server.")

	return carFiles, nil
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

func CreateCarFilesDescFromGoCarManifest(confCar *model.ConfCar, srcFileDir, carFileDir string) ([]*libmodel.FileDesc, error) {
	if confCar == nil {
		err := fmt.Errorf("parameter confCar is nil")
		logs.GetLogger().Error(err)
		return nil, err
	}

	manifestFilename := "manifest.csv"
	lines, err := utils.ReadAllLines(carFileDir, manifestFilename)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	carFiles := []*libmodel.FileDesc{}

	lotusClient, err := lotus.LotusGetClient(confCar.LotusClientApiUrl, confCar.LotusClientAccessToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

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

		carFile := libmodel.FileDesc{}
		carFile.PayloadCid = fields[0]
		carFile.CarFileName = carFile.PayloadCid + ".car"
		carFile.CarFilePath = filepath.Join(carFileDir, carFile.CarFileName)
		carFile.PieceCid = fields[2]
		carFile.CarFileSize = utils.GetInt64FromStr(fields[3])

		pieceCid := lotusClient.LotusClientCalcCommP(carFile.CarFilePath)
		if pieceCid == nil {
			err := fmt.Errorf("failed to generate piece cid")
			logs.GetLogger().Error(err)
			return nil, err
		}

		carFile.PieceCid = *pieceCid
		dataCid, err := lotusClient.LotusClientImport(carFile.CarFilePath, true)
		if err != nil {
			err := fmt.Errorf("failed to import car file")
			logs.GetLogger().Error(err)
			return nil, err
		}

		carFile.PayloadCid = *dataCid

		carFileDetail := fields[4]
		for i := 5; i < len(fields); i++ {
			carFileDetail = carFileDetail + "," + fields[i]
		}

		manifestDetail := ManifestDetail{}
		err = json.Unmarshal([]byte(carFileDetail), &manifestDetail)
		if err != nil {
			logs.GetLogger().Error("Failed to parse: ", carFileDetail)
			return nil, err
		}

		carFile.SourceFileName = manifestDetail.Link[0].Name
		carFile.SourceFilePath = filepath.Join(srcFileDir, carFile.SourceFileName)
		carFile.SourceFileSize = int64(manifestDetail.Link[0].Size)

		if confCar.GenerateMd5 {
			if utils.IsFileExistsFullPath(carFile.SourceFilePath) {
				srcFileMd5, err := checksum.MD5sum(carFile.SourceFilePath)
				if err != nil {
					logs.GetLogger().Error(err)
					return nil, err
				}
				carFile.SourceFileMd5 = srcFileMd5
			}

			carFileMd5, err := checksum.MD5sum(carFile.CarFilePath)
			if err != nil {
				logs.GetLogger().Error(err)
				return nil, err
			}
			carFile.CarFileMd5 = carFileMd5
		}

		carFiles = append(carFiles, &carFile)
	}

	_, err = WriteFileDescsToJsonFile(carFiles, carFileDir, constants.JSON_FILE_NAME_CAR_UPLOAD)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return carFiles, nil
}
