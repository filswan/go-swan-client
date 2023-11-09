package command

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/codingsince1985/checksum"
	"github.com/filswan/go-swan-client/config"
	"github.com/filswan/go-swan-lib/client"
	"github.com/filswan/go-swan-lib/client/lotus"
	"github.com/filswan/go-swan-lib/logs"
	libmodel "github.com/filswan/go-swan-lib/model"
	"github.com/filswan/go-swan-lib/utils"
	"github.com/google/uuid"
)

type CmdIpfsCmdCar struct {
	LotusClientApiUrl      string //required
	LotusClientAccessToken string //required
	OutputDir              string //required
	InputDir               string //required
	GenerateMd5            bool   //required
	ImportFlag             bool
}

func GetCmdIpfsCmdCar(inputDir string, outputDir *string, importFlag bool) *CmdIpfsCmdCar {
	cmdIpfsCmdCar := &CmdIpfsCmdCar{
		LotusClientApiUrl:      config.GetConfig().Lotus.ClientApiUrl,
		LotusClientAccessToken: config.GetConfig().Lotus.ClientAccessToken,
		InputDir:               inputDir,
		GenerateMd5:            config.GetConfig().Sender.GenerateMd5,
		ImportFlag:             importFlag,
	}

	if !utils.IsStrEmpty(outputDir) {
		cmdIpfsCmdCar.OutputDir = *outputDir
	} else {
		cmdIpfsCmdCar.OutputDir = filepath.Join(*outputDir, time.Now().Format("2006-01-02_15:04:05")) + "_" + uuid.NewString()
	}

	return cmdIpfsCmdCar
}

func CreateIpfsCmdCarFilesByConfig(inputDir string, outputDir *string, importFlag bool) ([]*libmodel.FileDesc, error) {
	cmdIpfsCmdCar := GetCmdIpfsCmdCar(inputDir, outputDir, importFlag)
	fileDescs, err := cmdIpfsCmdCar.CreateIpfsCmdCarFiles()
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return fileDescs, nil
}

func (cmdIpfsCmdCar *CmdIpfsCmdCar) CreateIpfsCmdCarFiles() ([]*libmodel.FileDesc, error) {
	err := utils.CheckDirExists(cmdIpfsCmdCar.InputDir, DIR_NAME_INPUT)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	err = utils.CreateDirIfNotExists(cmdIpfsCmdCar.OutputDir, DIR_NAME_OUTPUT)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	srcFileSize, err := utils.GetFilesSize(cmdIpfsCmdCar.InputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if *srcFileSize == 0 {
		err := fmt.Errorf("no files with contents under directory:%s", cmdIpfsCmdCar.InputDir)
		logs.GetLogger().Error(err)
		return nil, err
	}

	logs.GetLogger().Info("Creating car file for ", cmdIpfsCmdCar.InputDir)
	carFileName := filepath.Base(cmdIpfsCmdCar.InputDir) + ".car"
	carFilePath := filepath.Join(cmdIpfsCmdCar.OutputDir, carFileName)
	ipfsCmdCarCmd := fmt.Sprintf("ipfs-car --pack %s --output %s", cmdIpfsCmdCar.InputDir, carFilePath)
	result, err := client.ExecOsCmd2Screen(ipfsCmdCarCmd, true)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if strings.Contains(result, "Error") {
		err := fmt.Errorf(result)
		logs.GetLogger().Error(err)
		return nil, err
	}

	fileDesc := libmodel.FileDesc{}
	fileDesc.SourceFileName = filepath.Base(cmdIpfsCmdCar.InputDir)
	fileDesc.SourceFilePath = cmdIpfsCmdCar.InputDir
	fileDesc.SourceFileSize = *srcFileSize
	fileDesc.CarFileName = carFileName
	fileDesc.CarFileUrl = fileDesc.CarFileName
	fileDesc.CarFilePath = carFilePath

	if cmdIpfsCmdCar.ImportFlag {
		lotusClient, err := lotus.LotusGetClient(cmdIpfsCmdCar.LotusClientApiUrl, cmdIpfsCmdCar.LotusClientAccessToken)
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

		dataCid, err := lotusClient.LotusClientImport(fileDesc.CarFilePath, true)
		if err != nil {
			err := fmt.Errorf("failed to import car file to lotus client")
			logs.GetLogger().Error(err)
			return nil, err
		}
		fileDesc.PayloadCid = *dataCid
	} else {
		dataCid, pieceCid, _, err := CalculateValueByCarFile(fileDesc.CarFilePath, true, true)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}
		fileDesc.PayloadCid = dataCid
		fileDesc.PieceCid = pieceCid
	}

	fileDesc.CarFileSize = utils.GetFileSize(fileDesc.CarFilePath)

	if cmdIpfsCmdCar.GenerateMd5 {
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

	fileDescs := []*libmodel.FileDesc{
		&fileDesc,
	}

	_, err = WriteCarFilesToFiles(fileDescs, cmdIpfsCmdCar.OutputDir, JSON_FILE_NAME_CAR_UPLOAD, CSV_FILE_NAME_CAR_UPLOAD)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	logs.GetLogger().Info(len(fileDescs), " car files have been created to directory:", cmdIpfsCmdCar.OutputDir)
	logs.GetLogger().Info("Please upload car files to web server or ipfs server.")

	return fileDescs, nil
}
