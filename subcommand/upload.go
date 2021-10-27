package subcommand

import (
	"fmt"

	"github.com/filswan/go-swan-lib/client"
	libmodel "github.com/filswan/go-swan-lib/model"

	"github.com/filswan/go-swan-client/model"
	"github.com/filswan/go-swan-lib/constants"
	"github.com/filswan/go-swan-lib/logs"
	"github.com/filswan/go-swan-lib/utils"
)

func UploadCarFiles(confUpload *model.ConfUpload) ([]*libmodel.FileDesc, error) {
	err := CheckInputDir(confUpload.InputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if confUpload.StorageServerType == constants.STORAGE_SERVER_TYPE_WEB_SERVER {
		logs.GetLogger().Info("Please upload car files to web server manually.")
		return nil, nil
	}

	carFiles := ReadCarFilesFromJsonFile(confUpload.InputDir, constants.JSON_FILE_NAME_BY_CAR)
	if carFiles == nil {
		err := fmt.Errorf("failed to read:%s", confUpload.InputDir)
		logs.GetLogger().Error(err)
		return nil, err
	}

	for _, carFile := range carFiles {
		logs.GetLogger().Info("Uploading car file:", carFile.CarFilePath)
		carFileHash, err := client.IpfsUploadCarFile(carFile.CarFilePath)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}

		carFileUrl := utils.UrlJoin(confUpload.IpfsServerDownloadUrlPrefix, *carFileHash)
		carFile.CarFileUrl = &carFileUrl
		logs.GetLogger().Info("Car file: ", carFile.CarFileName, " uploaded to: ", carFile.CarFileUrl)
	}

	err = WriteCarFilesToFiles(carFiles, confUpload.InputDir, constants.JSON_FILE_NAME_BY_UPLOAD, constants.CSV_FILE_NAME_BY_UPLOAD)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return carFiles, nil
}
