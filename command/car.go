package command

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/codingsince1985/checksum"
	"github.com/filswan/go-swan-client/config"
	"github.com/filswan/go-swan-lib/client/lotus"
	"github.com/filswan/go-swan-lib/logs"
	libmodel "github.com/filswan/go-swan-lib/model"
	"github.com/filswan/go-swan-lib/utils"
	"github.com/google/uuid"
)

type CmdCar struct {
	LotusClientApiUrl      string //required
	LotusClientAccessToken string //required
	OutputDir              string //required
	InputDir               string //required
	GenerateMd5            bool   //required
	ImportFlag             bool
}

func GetCmdCar(inputDir string, outputDir *string, importFlag bool) *CmdCar {
	cmdCar := &CmdCar{
		LotusClientApiUrl:      config.GetConfig().Lotus.ClientApiUrl,
		LotusClientAccessToken: config.GetConfig().Lotus.ClientAccessToken,
		InputDir:               inputDir,
		GenerateMd5:            config.GetConfig().Sender.GenerateMd5,
		ImportFlag:             importFlag,
	}

	if !utils.IsStrEmpty(outputDir) {
		cmdCar.OutputDir = *outputDir
	} else {
		cmdCar.OutputDir = filepath.Join(*outputDir, time.Now().Format("2006-01-02_15:04:05")) + "_" + uuid.NewString()
	}

	return cmdCar
}

func CreateCarFilesByConfig(inputDir string, outputDir *string, importFlag bool) ([]*libmodel.FileDesc, error) {
	cmdCar := GetCmdCar(inputDir, outputDir, importFlag)
	fileDescs, err := cmdCar.CreateCarFiles()
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return fileDescs, nil
}

func (cmdCar *CmdCar) CreateCarFiles() ([]*libmodel.FileDesc, error) {
	err := utils.CheckDirExists(cmdCar.InputDir, DIR_NAME_INPUT)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	err = utils.CreateDirIfNotExists(cmdCar.OutputDir, DIR_NAME_OUTPUT)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	srcFiles, err := ioutil.ReadDir(cmdCar.InputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	fileDescs := []*libmodel.FileDesc{}

	lotusClient, err := lotus.LotusGetClient(cmdCar.LotusClientApiUrl, cmdCar.LotusClientAccessToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	for _, srcFile := range srcFiles {
		fileDesc := libmodel.FileDesc{}
		fileDesc.SourceFileName = srcFile.Name()
		fileDesc.SourceFilePath = filepath.Join(cmdCar.InputDir, fileDesc.SourceFileName)
		fileDesc.SourceFileSize = srcFile.Size()
		fileDesc.CarFileName = fileDesc.SourceFileName + ".car"
		fileDesc.CarFileUrl = fileDesc.CarFileName
		fileDesc.CarFilePath = filepath.Join(cmdCar.OutputDir, fileDesc.CarFileName)
		logs.GetLogger().Info("Creating car file ", fileDesc.CarFilePath, " for ", fileDesc.SourceFilePath)

		err := lotusClient.LotusClientGenCar(fileDesc.SourceFilePath, fileDesc.CarFilePath, false)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}

		pieceCid, err := lotusClient.LotusClientCalcCommP(fileDesc.CarFilePath)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}

		fileDesc.PieceCid = *pieceCid

		if cmdCar.ImportFlag {
			dataCid, err := lotusClient.LotusClientImport(fileDesc.CarFilePath, true)
			if err != nil {
				err := fmt.Errorf("failed to import car file")
				logs.GetLogger().Error(err)
				return nil, err
			}

			if dataCid == nil {
				err := fmt.Errorf("failed to generate data cid for: %s", fileDesc.CarFilePath)
				logs.GetLogger().Error(err)
				return nil, err
			}

			fileDesc.PayloadCid = *dataCid
		} else {
			dataCid, _, _, err := CalculateValueByCarFile(fileDesc.CarFilePath, true, false)
			if err != nil {
				err := fmt.Errorf("failed to generate data cid for: %s", fileDesc.CarFilePath)
				logs.GetLogger().Error(err)
				return nil, err
			}
			fileDesc.PayloadCid = dataCid
		}

		fileDesc.CarFileSize = utils.GetFileSize(fileDesc.CarFilePath)

		if cmdCar.GenerateMd5 {
			srcFileMd5, err := checksum.MD5sum(fileDesc.SourceFilePath)
			if err != nil {
				logs.GetLogger().Error(err)
				return nil, err
			}
			fileDesc.SourceFileMd5 = srcFileMd5

			carFileMd5, err := checksum.MD5sum(fileDesc.CarFilePath)
			if err != nil {
				logs.GetLogger().Error(err)
				return nil, err
			}
			fileDesc.CarFileMd5 = carFileMd5
		}

		fileDescs = append(fileDescs, &fileDesc)
		logs.GetLogger().Info("Car file ", fileDesc.CarFilePath, " created")
	}

	logs.GetLogger().Info(len(fileDescs), " car files have been created to directory:", cmdCar.OutputDir)

	_, err = WriteCarFilesToFiles(fileDescs, cmdCar.OutputDir, JSON_FILE_NAME_CAR_UPLOAD, CSV_FILE_NAME_CAR_UPLOAD)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	logs.GetLogger().Info("Please upload car files to web server or ipfs server.")

	return fileDescs, nil
}
