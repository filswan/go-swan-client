package command

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/codingsince1985/checksum"
	"github.com/filswan/go-swan-client/config"
	"github.com/filswan/go-swan-lib/client/ipfs"
	"github.com/filswan/go-swan-lib/client/lotus"
	"github.com/filswan/go-swan-lib/logs"
	libmodel "github.com/filswan/go-swan-lib/model"
	"github.com/filswan/go-swan-lib/utils"
)

type CmdIpfsCar struct {
	LotusClientApiUrl         string //required
	LotusClientAccessToken    string //required
	OutputDir                 string //required
	InputDir                  string //required
	GenerateMd5               bool   //required
	IpfsServerUploadUrlPrefix string //required
}

func GetCmdIpfsCar(inputDir string, outputDir *string) *CmdIpfsCar {
	cmdIpfsCar := &CmdIpfsCar{
		LotusClientApiUrl:         config.GetConfig().Lotus.ClientApiUrl,
		LotusClientAccessToken:    config.GetConfig().Lotus.ClientAccessToken,
		OutputDir:                 filepath.Join(config.GetConfig().Sender.OutputDir, time.Now().Format("2006-01-02_15:04:05")),
		InputDir:                  inputDir,
		GenerateMd5:               config.GetConfig().Sender.GenerateMd5,
		IpfsServerUploadUrlPrefix: config.GetConfig().IpfsServer.UploadUrlPrefix,
	}

	if !utils.IsStrEmpty(outputDir) {
		cmdIpfsCar.OutputDir = *outputDir
	}

	return cmdIpfsCar
}

func CreateIpfsCarFilesByConfig(inputDir string, outputDir *string) ([]*libmodel.FileDesc, error) {
	cmdIpfsCar := GetCmdIpfsCar(inputDir, outputDir)
	fileDescs, err := cmdIpfsCar.CreateIpfsCarFiles()
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return fileDescs, nil
}

func (cmdIpfsCar *CmdIpfsCar) CreateIpfsCarFiles() ([]*libmodel.FileDesc, error) {
	if cmdIpfsCar.IpfsServerUploadUrlPrefix == "" {
		err := fmt.Errorf("IpfsServerUploadUrlPrefix is required")
		logs.GetLogger().Error(err)
		return nil, err
	}

	err := utils.CheckDirExists(cmdIpfsCar.InputDir, DIR_NAME_INPUT)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	err = utils.CreateDirIfNotExists(cmdIpfsCar.OutputDir, DIR_NAME_OUTPUT)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	srcFiles, err := ioutil.ReadDir(cmdIpfsCar.InputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if len(srcFiles) == 0 {
		err := fmt.Errorf("no files under directory:%s", cmdIpfsCar.InputDir)
		logs.GetLogger().Error(err)
		return nil, err
	}

	lotusClient, err := lotus.LotusGetClient(cmdIpfsCar.LotusClientApiUrl, cmdIpfsCar.LotusClientAccessToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	logs.GetLogger().Info("Creating car file for ", cmdIpfsCar.InputDir)
	srcFileCids := []string{}
	for _, srcFile := range srcFiles {
		srcFilePath := filepath.Join(cmdIpfsCar.InputDir, srcFile.Name())
		srcFileCid, err := ipfs.IpfsUploadFileByWebApi(utils.UrlJoin(cmdIpfsCar.IpfsServerUploadUrlPrefix, "api/v0/add?stream-channels=true&pin=true"), srcFilePath)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}

		srcFileCids = append(srcFileCids, *srcFileCid)
	}
	carFileDataCid, err := ipfs.MergeFiles2CarFile(cmdIpfsCar.IpfsServerUploadUrlPrefix, srcFileCids)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	carFileName := *carFileDataCid + ".car"
	carFilePath := filepath.Join(cmdIpfsCar.OutputDir, carFileName)
	err = ipfs.Export2CarFile(cmdIpfsCar.IpfsServerUploadUrlPrefix, *carFileDataCid, carFilePath)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	carFiles := []*libmodel.FileDesc{}
	carFile := libmodel.FileDesc{}
	carFile.CarFileName = carFileName
	carFile.CarFilePath = carFilePath

	pieceCid, err := lotusClient.LotusClientCalcCommP(carFile.CarFilePath)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	carFile.PieceCid = *pieceCid

	dataCid, err := lotusClient.LotusClientImport(carFile.CarFilePath, true)
	if err != nil {
		err := fmt.Errorf("failed to import car file to lotus client")
		logs.GetLogger().Error(err)
		return nil, err
	}

	carFile.PayloadCid = *dataCid

	carFile.CarFileSize = utils.GetFileSize(carFile.CarFilePath)

	if cmdIpfsCar.GenerateMd5 {
		carFileMd5, err := checksum.MD5sum(carFile.CarFilePath)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}
		carFile.CarFileMd5 = carFileMd5
	}

	carFiles = append(carFiles, &carFile)

	_, err = WriteFileDescsToJsonFile(carFiles, cmdIpfsCar.OutputDir, JSON_FILE_NAME_CAR_UPLOAD)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	logs.GetLogger().Info(len(carFiles), " car files have been created to directory:", cmdIpfsCar.OutputDir)
	logs.GetLogger().Info("Please upload car files to web server or ipfs server.")

	return carFiles, nil
}
