package command

import (
	"fmt"

	"github.com/filswan/go-swan-lib/client/ipfs"
	libmodel "github.com/filswan/go-swan-lib/model"

	"github.com/filswan/go-swan-client/config"

	libconstants "github.com/filswan/go-swan-lib/constants"
	"github.com/filswan/go-swan-lib/logs"
	"github.com/filswan/go-swan-lib/utils"
)

type CmdUpload struct {
	StorageServerType           string //required
	IpfsServerDownloadUrlPrefix string //required only when upload to ipfs server
	IpfsServerUploadUrlPrefix   string //required only when upload to ipfs server
	OutputDir                   string //invalid
	InputDir                    string //required
}

func GetCmdUpload(inputDir string) *CmdUpload {
	cmdUpload := &CmdUpload{
		StorageServerType:           config.GetConfig().Main.StorageServerType,
		IpfsServerDownloadUrlPrefix: config.GetConfig().IpfsServer.DownloadUrlPrefix,
		IpfsServerUploadUrlPrefix:   config.GetConfig().IpfsServer.UploadUrlPrefix,
		OutputDir:                   inputDir,
		InputDir:                    inputDir,
	}

	return cmdUpload
}

func UploadCarFilesByConfig(inputDir string) ([]*libmodel.FileDesc, error) {
	cmdUpload := GetCmdUpload(inputDir)

	fileDescs, err := cmdUpload.UploadCarFiles()
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return fileDescs, nil
}

func (cmdUpload *CmdUpload) UploadCarFiles() ([]*libmodel.FileDesc, error) {
	err := CheckInputDir(cmdUpload.InputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if cmdUpload.StorageServerType == libconstants.STORAGE_SERVER_TYPE_WEB_SERVER {
		logs.GetLogger().Info("Please upload car files to web server manually.")
		return nil, nil
	}

	carFiles := ReadFileDescsFromJsonFile(cmdUpload.InputDir, JSON_FILE_NAME_CAR_UPLOAD)
	if carFiles == nil {
		err := fmt.Errorf("failed to read:%s", cmdUpload.InputDir)
		logs.GetLogger().Error(err)
		return nil, err
	}

	for _, carFile := range carFiles {
		uploadUrl := utils.UrlJoin(cmdUpload.IpfsServerUploadUrlPrefix, "api/v0/add?stream-channels=true&pin=true")
		logs.GetLogger().Info("Uploading car file:", carFile.CarFilePath, " to:", uploadUrl)
		carFileHash, err := ipfs.IpfsUploadFileByWebApi(uploadUrl, carFile.CarFilePath)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}

		carFileUrl := utils.UrlJoin(cmdUpload.IpfsServerDownloadUrlPrefix, "ipfs", *carFileHash)
		carFile.CarFileUrl = carFileUrl
		logs.GetLogger().Info("Car file: ", carFile.CarFileName, " uploaded to: ", carFile.CarFileUrl)
	}

	_, err = WriteFileDescsToJsonFile(carFiles, cmdUpload.InputDir, JSON_FILE_NAME_CAR_UPLOAD)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	logs.GetLogger().Info("Please create a task for your car file(s)")

	return carFiles, nil
}
