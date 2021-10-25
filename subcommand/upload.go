package subcommand

import (
	"fmt"

	"go-swan-client/common/client"
	"go-swan-client/common/constants"
	"go-swan-client/common/utils"
	"go-swan-client/logs"
	"go-swan-client/model"
)

func UploadCarFiles(confUpload *model.ConfUpload) error {
	err := CheckInputDir(confUpload.InputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	if confUpload.StorageServerType == constants.STORAGE_SERVER_TYPE_WEB_SERVER {
		logs.GetLogger().Info("Please upload car files to web server manually.")
		return nil
	}

	carFiles := ReadCarFilesFromJsonFile(confUpload.InputDir, constants.JSON_FILE_NAME_BY_CAR)
	if carFiles == nil {
		err := fmt.Errorf("failed to read:%s", confUpload.InputDir)
		logs.GetLogger().Error(err)
		return err
	}

	for _, carFile := range carFiles {
		logs.GetLogger().Info("Uploading car file:", carFile.CarFilePath)
		carFileHash, err := client.IpfsUploadCarFile(carFile.CarFilePath)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}

		carFileUrl := utils.UrlJoin(confUpload.IpfsServerDownloadUrlPrefix, *carFileHash)
		carFile.CarFileUrl = &carFileUrl
		logs.GetLogger().Info("Car file: ", carFile.CarFileName, " uploaded to: ", carFile.CarFileUrl)
	}

	err = WriteCarFilesToFiles(carFiles, confUpload.InputDir, constants.JSON_FILE_NAME_BY_UPLOAD, constants.CSV_FILE_NAME_BY_UPLOAD)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}
