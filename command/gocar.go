package command

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	"github.com/filswan/go-swan-client/config"

	"github.com/filswan/go-swan-lib/logs"

	"github.com/filswan/go-swan-lib/utils"

	"github.com/codingsince1985/checksum"
	"github.com/filedrive-team/go-graphsplit"
	"github.com/filswan/go-swan-lib/client/lotus"
	libmodel "github.com/filswan/go-swan-lib/model"
)

type CmdGoCar struct {
	LotusClientApiUrl         string //required
	LotusClientAccessToken    string //required
	OutputDir                 string //required
	InputDir                  string //required
	GocarFileSizeLimit        int64  //required only when creating gocar file(s)
	GenerateMd5               bool   //required
	IpfsServerUploadUrlPrefix string //required only when creating ipfs car file
}

func GetCmdGoCar(inputDir string, outputDir *string) *CmdGoCar {
	cmdGoCar := &CmdGoCar{
		LotusClientApiUrl:         config.GetConfig().Lotus.ClientApiUrl,
		LotusClientAccessToken:    config.GetConfig().Lotus.ClientAccessToken,
		OutputDir:                 filepath.Join(config.GetConfig().Sender.OutputDir, time.Now().Format("2006-01-02_15:04:05")),
		InputDir:                  inputDir,
		GocarFileSizeLimit:        config.GetConfig().Sender.GocarFileSizeLimit,
		GenerateMd5:               config.GetConfig().Sender.GenerateMd5,
		IpfsServerUploadUrlPrefix: config.GetConfig().IpfsServer.UploadUrlPrefix,
	}

	if outputDir != nil && len(*outputDir) != 0 {
		cmdGoCar.OutputDir = *outputDir
	}

	return cmdGoCar
}

func CreateGoCarFilesByConfig(inputDir string, outputDir *string) ([]*libmodel.FileDesc, error) {
	cmdGoCar := GetCmdGoCar(inputDir, outputDir)
	fileDescs, err := cmdGoCar.CreateGoCarFiles()
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return fileDescs, nil
}

func (cmdGoCar *CmdGoCar) CreateGoCarFiles() ([]*libmodel.FileDesc, error) {
	err := CheckInputDir(cmdGoCar.InputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	err = CreateOutputDir(cmdGoCar.OutputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	sliceSize := cmdGoCar.GocarFileSizeLimit
	if sliceSize <= 0 {
		err := fmt.Errorf("gocar file size limit is too smal")
		logs.GetLogger().Error(err)
		return nil, err
	}

	srcFiles, err := ioutil.ReadDir(cmdGoCar.InputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	carDir := cmdGoCar.OutputDir
	for _, srcFile := range srcFiles {
		parentPath := filepath.Join(cmdGoCar.InputDir, srcFile.Name())
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
	carFiles, err := cmdGoCar.CreateCarFilesDescFromGoCarManifest(cmdGoCar.InputDir, carDir)
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

func (cmdGoCar *CmdGoCar) CreateCarFilesDescFromGoCarManifest(srcFileDir, carFileDir string) ([]*libmodel.FileDesc, error) {
	manifestFilename := "manifest.csv"
	lines, err := utils.ReadAllLines(carFileDir, manifestFilename)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	carFiles := []*libmodel.FileDesc{}

	lotusClient, err := lotus.LotusGetClient(cmdGoCar.LotusClientApiUrl, cmdGoCar.LotusClientAccessToken)
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

		if cmdGoCar.GenerateMd5 {
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

	_, err = WriteFileDescsToJsonFile(carFiles, carFileDir, JSON_FILE_NAME_CAR_UPLOAD)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return carFiles, nil
}
