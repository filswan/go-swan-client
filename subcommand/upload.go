package subcommand

import (
	"fmt"

	"github.com/filswan/go-swan-lib/client/ipfs"
	libmodel "github.com/filswan/go-swan-lib/model"

	"github.com/filswan/go-swan-client/common/constants"
	"github.com/filswan/go-swan-client/model"

	libconstants "github.com/filswan/go-swan-lib/constants"
	"github.com/filswan/go-swan-lib/logs"
	"github.com/filswan/go-swan-lib/utils"
)

func UploadCarFiles(confUpload *model.ConfUpload) ([]*libmodel.FileDesc, error) {
	err := CheckInputDir(confUpload.InputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if confUpload.StorageServerType == libconstants.STORAGE_SERVER_TYPE_WEB_SERVER {
		logs.GetLogger().Info("Please upload car files to web server manually.")
		return nil, nil
	}

	carFiles := ReadCarFilesFromJsonFile(confUpload.InputDir, constants.JSON_FILE_NAME_CAR_UPLOAD)
	if carFiles == nil {
		err := fmt.Errorf("failed to read:%s", confUpload.InputDir)
		logs.GetLogger().Error(err)
		return nil, err
	}

	for _, carFile := range carFiles {
		uploadUrl := utils.UrlJoin(confUpload.IpfsServerUploadUrlPrefix, "api/v0/add?stream-channels=true&pin=true")
		logs.GetLogger().Info("Uploading car file:", carFile.CarFilePath, " to:", uploadUrl)
		carFileHash, err := ipfs.IpfsUploadFileByWebApi(uploadUrl, carFile.CarFilePath)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}

		carFileUrl := utils.UrlJoin(confUpload.IpfsServerDownloadUrlPrefix, "ipfs", *carFileHash)
		carFile.CarFileUrl = carFileUrl
		logs.GetLogger().Info("Car file: ", carFile.CarFileName, " uploaded to: ", carFile.CarFileUrl)
	}

	_, err = WriteCarFilesToJsonFile(carFiles, confUpload.InputDir, constants.JSON_FILE_NAME_CAR_UPLOAD)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	logs.GetLogger().Info("Please create a task for your car file(s)")

	return carFiles, nil
}
