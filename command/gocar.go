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
	LotusClientApiUrl      string //required
	LotusClientAccessToken string //required
	OutputDir              string //required
	InputDir               string //required
	GenerateMd5            bool   //required
	GocarFileSizeLimit     int64  //required
	GocarFolderBased       bool   //required
}

func GetCmdGoCar(inputDir string, outputDir *string) *CmdGoCar {
	cmdGoCar := &CmdGoCar{
		LotusClientApiUrl:      config.GetConfig().Lotus.ClientApiUrl,
		LotusClientAccessToken: config.GetConfig().Lotus.ClientAccessToken,
		OutputDir:              filepath.Join(config.GetConfig().Sender.OutputDir, time.Now().Format("2006-01-02_15:04:05")),
		InputDir:               inputDir,
		GocarFileSizeLimit:     config.GetConfig().Sender.GocarFileSizeLimit,
		GenerateMd5:            config.GetConfig().Sender.GenerateMd5,
		GocarFolderBased:       config.GetConfig().Sender.GocarFolderBased,
	}

	if !utils.IsStrEmpty(outputDir) {
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
	err := utils.CheckDirExists(cmdGoCar.InputDir, DIR_NAME_INPUT)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	err = utils.CreateDirIfNotExists(cmdGoCar.OutputDir, DIR_NAME_OUTPUT)
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
	Emptyctx := context.Background()
	parallel := 4
	cb := graphsplit.CommPCallback(carDir)

	if cmdGoCar.GocarFolderBased {
		parentPath := cmdGoCar.InputDir
		targetPath := parentPath
		graphName := filepath.Base(parentPath)

		logs.GetLogger().Info("Creating car file for ", parentPath)
		err = graphsplit.Chunk(Emptyctx, sliceSize, parentPath, targetPath, carDir, graphName, parallel, cb)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}
		logs.GetLogger().Info("Car file for ", parentPath, " created")
	} else {
		for _, srcFile := range srcFiles {
			parentPath := filepath.Join(cmdGoCar.InputDir, srcFile.Name())
			targetPath := parentPath
			graphName := srcFile.Name()

			logs.GetLogger().Info("Creating car file for ", parentPath)
			err = graphsplit.Chunk(Emptyctx, sliceSize, parentPath, targetPath, carDir, graphName, parallel, cb)
			if err != nil {
				logs.GetLogger().Error(err)
				return nil, err
			}
			logs.GetLogger().Info("Car file for ", parentPath, " created")
		}
	}
	fileDescs, err := cmdGoCar.createFilesDescFromManifest()
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	logs.GetLogger().Info(len(fileDescs), " car files have been created to directory:", carDir)
	logs.GetLogger().Info("Please upload car files to web server or ipfs server.")

	return fileDescs, nil
}

type ManifestDetail struct {
	Name string
	Hash string
	Size int
	Link []struct {
		Name string
		Hash string
		Size int64
	}
}

func (cmdGoCar *CmdGoCar) createFilesDescFromManifest() ([]*libmodel.FileDesc, error) {
	manifestFilename := "manifest.csv"
	lines, err := utils.ReadAllLines(cmdGoCar.OutputDir, manifestFilename)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	lotusClient, err := lotus.LotusGetClient(cmdGoCar.LotusClientApiUrl, cmdGoCar.LotusClientAccessToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	fileDescs := []*libmodel.FileDesc{}
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

		fileDesc := libmodel.FileDesc{}
		fileDesc.PayloadCid = fields[0]
		fileDesc.CarFileName = fileDesc.PayloadCid + ".car"
		fileDesc.CarFilePath = filepath.Join(cmdGoCar.OutputDir, fileDesc.CarFileName)
		fileDesc.PieceCid = fields[2]
		fileDesc.CarFileSize = utils.GetInt64FromStr(fields[3])

		pieceCid, err := lotusClient.LotusClientCalcCommP(fileDesc.CarFilePath)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}

		fileDesc.PieceCid = *pieceCid
		dataCid, err := lotusClient.LotusClientImport(fileDesc.CarFilePath, true)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}

		fileDesc.PayloadCid = *dataCid

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

		if cmdGoCar.GocarFolderBased {
			fileDesc.SourceFileName = filepath.Base(cmdGoCar.InputDir)
			fileDesc.SourceFilePath = cmdGoCar.InputDir
			for _, link := range manifestDetail.Link {
				fileDesc.SourceFileSize = fileDesc.SourceFileSize + link.Size
			}
		} else {
			fileDesc.SourceFileName = manifestDetail.Link[0].Name
			fileDesc.SourceFilePath = filepath.Join(cmdGoCar.InputDir, fileDesc.SourceFileName)
			fileDesc.SourceFileSize = int64(manifestDetail.Link[0].Size)
		}

		if cmdGoCar.GenerateMd5 {
			if utils.IsFileExistsFullPath(fileDesc.SourceFilePath) {
				srcFileMd5, err := checksum.MD5sum(fileDesc.SourceFilePath)
				if err != nil {
					logs.GetLogger().Error(err)
					return nil, err
				}
				fileDesc.SourceFileMd5 = srcFileMd5
			}

			carFileMd5, err := checksum.MD5sum(fileDesc.CarFilePath)
			if err != nil {
				logs.GetLogger().Error(err)
				return nil, err
			}
			fileDesc.CarFileMd5 = carFileMd5
		}

		fileDescs = append(fileDescs, &fileDesc)
	}

	_, err = WriteFileDescsToJsonFile(fileDescs, cmdGoCar.OutputDir, JSON_FILE_NAME_CAR_UPLOAD)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return fileDescs, nil
}
