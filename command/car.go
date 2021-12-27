package command

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/filswan/go-swan-client/config"

	"github.com/filswan/go-swan-lib/client/lotus"
	"github.com/filswan/go-swan-lib/logs"

	"github.com/filswan/go-swan-lib/utils"

	"github.com/codingsince1985/checksum"
	libmodel "github.com/filswan/go-swan-lib/model"
)

type CmdCar struct {
	LotusClientApiUrl      string //required
	LotusClientAccessToken string //required
	OutputDir              string //required
	InputDir               string //required
	GenerateMd5            bool   //required
}

func GetCmdCar(inputDir string, outputDir *string) *CmdCar {
	cmdCar := &CmdCar{
		LotusClientApiUrl:      config.GetConfig().Lotus.ClientApiUrl,
		LotusClientAccessToken: config.GetConfig().Lotus.ClientAccessToken,
		OutputDir:              filepath.Join(config.GetConfig().Sender.OutputDir, time.Now().Format("2006-01-02_15:04:05")),
		InputDir:               inputDir,
		GenerateMd5:            config.GetConfig().Sender.GenerateMd5,
	}

	if outputDir != nil && len(*outputDir) != 0 {
		cmdCar.OutputDir = *outputDir
	}

	return cmdCar
}

func CreateCarFilesByConfig(inputDir string, outputDir *string) ([]*libmodel.FileDesc, error) {
	cmdCar := GetCmdCar(inputDir, outputDir)
	fileDescs, err := cmdCar.CreateCarFiles()
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return fileDescs, nil
}

func (cmdCar *CmdCar) CreateCarFiles() ([]*libmodel.FileDesc, error) {
	err := CheckInputDir(cmdCar.InputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	err = CreateOutputDir(cmdCar.OutputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	srcFiles, err := ioutil.ReadDir(cmdCar.InputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	carFiles := []*libmodel.FileDesc{}

	lotusClient, err := lotus.LotusGetClient(cmdCar.LotusClientApiUrl, cmdCar.LotusClientAccessToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	for _, srcFile := range srcFiles {
		carFile := libmodel.FileDesc{}
		carFile.SourceFileName = srcFile.Name()
		carFile.SourceFilePath = filepath.Join(cmdCar.InputDir, carFile.SourceFileName)
		carFile.SourceFileSize = srcFile.Size()
		carFile.CarFileName = carFile.SourceFileName + ".car"
		carFile.CarFilePath = filepath.Join(cmdCar.OutputDir, carFile.CarFileName)
		logs.GetLogger().Info("Creating car file ", carFile.CarFilePath, " for ", carFile.SourceFilePath)

		err := lotusClient.LotusClientGenCar(carFile.SourceFilePath, carFile.CarFilePath, false)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}

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

		if dataCid == nil {
			err := fmt.Errorf("failed to generate data cid for: %s", carFile.CarFilePath)
			logs.GetLogger().Error(err)
			return nil, err
		}

		carFile.PayloadCid = *dataCid

		carFile.CarFileSize = utils.GetFileSize(carFile.CarFilePath)

		if cmdCar.GenerateMd5 {
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
		}

		carFiles = append(carFiles, &carFile)
		logs.GetLogger().Info("Car file ", carFile.CarFilePath, " created")
	}

	logs.GetLogger().Info(len(carFiles), " car files have been created to directory:", cmdCar.OutputDir)

	_, err = WriteFileDescsToJsonFile(carFiles, cmdCar.OutputDir, JSON_FILE_NAME_CAR_UPLOAD)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	logs.GetLogger().Info("Please upload car files to web server or ipfs server.")

	return carFiles, nil
}
