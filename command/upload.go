package command

import (
	"fmt"

	"github.com/filswan/go-swan-client/config"
	"github.com/filswan/go-swan-lib/client/ipfs"
	"github.com/filswan/go-swan-lib/logs"
	libmodel "github.com/filswan/go-swan-lib/model"
	"github.com/filswan/go-swan-lib/utils"
)

type CmdUpload struct {
	IpfsServerDownloadUrlPrefix string //required only when upload to ipfs server
	IpfsServerUploadUrlPrefix   string //required only when upload to ipfs server
	InputDir                    string //required
}

func GetCmdUpload(inputDir string) *CmdUpload {
	cmdUpload := &CmdUpload{
		IpfsServerDownloadUrlPrefix: config.GetConfig().IpfsServer.DownloadUrlPrefix,
		IpfsServerUploadUrlPrefix:   config.GetConfig().IpfsServer.UploadUrlPrefix,
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
	err := utils.CheckDirExists(cmdUpload.InputDir, DIR_NAME_INPUT)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	fileDescs, err := ReadFileDescsFromJsonFile(cmdUpload.InputDir, JSON_FILE_NAME_CAR_UPLOAD)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if fileDescs == nil {
		err := fmt.Errorf("failed to read:%s", cmdUpload.InputDir)
		logs.GetLogger().Error(err)
		return nil, err
	}
	uploadUrl := utils.UrlJoin(cmdUpload.IpfsServerUploadUrlPrefix, "api/v0/add?stream-channels=true&pin=true")
	for _, fileDesc := range fileDescs {
		logs.GetLogger().Info("Uploading car file:", fileDesc.CarFilePath, " to:", uploadUrl)
		carFileHash, err := ipfs.IpfsUploadFileByWebApi(uploadUrl, fileDesc.CarFilePath)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}

		carFileUrl := utils.UrlJoin(cmdUpload.IpfsServerDownloadUrlPrefix, "ipfs", *carFileHash)
		fileDesc.CarFileUrl = carFileUrl
		logs.GetLogger().Info("Car file: ", fileDesc.CarFileName, " uploaded to: ", fileDesc.CarFileUrl)
	}

	logs.GetLogger().Info(len(fileDescs), " car files have been uploaded to:", uploadUrl)

	_, err = WriteCarFilesToFiles(fileDescs, cmdUpload.InputDir, JSON_FILE_NAME_CAR_UPLOAD, CSV_FILE_NAME_CAR_UPLOAD)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	logs.GetLogger().Info("Please create a task for your car file(s)")

	return fileDescs, nil
}
