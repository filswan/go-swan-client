package subcommand

import (
	"fmt"

	"go-swan-client/common/client"
	"go-swan-client/common/constants"
	"go-swan-client/config"
	"go-swan-client/logs"
)

func UploadCarFiles(inputDir string) error {
	err := CheckInputDir(inputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	storageServerType := config.GetConfig().Main.StorageServerType
	if storageServerType == constants.STORAGE_SERVER_TYPE_WEB_SERVER {
		logs.GetLogger().Info("Please upload car files to web server manually.")
		return nil
	}

	//gatewayAddress := config.GetConfig().IpfsServer.GatewayAddress
	//words := strings.Split(gatewayAddress, "/")
	//if len(words) < 5 {
	//	err := fmt.Errorf("invalid gateway address:%s", gatewayAddress)
	//	logs.GetLogger().Error(err)
	//	return err
	//}
	//gatewayIp := words[2]
	//gatewayPort := words[4]

	carFiles := ReadCarFilesFromJsonFile(inputDir, constants.JSON_FILE_NAME_BY_CAR)
	if carFiles == nil {
		err := fmt.Errorf("failed to read:%s", inputDir)
		logs.GetLogger().Error(err)
		return err
	}
	uploadUrl := config.GetConfig().IpfsServer.UploadUrl
	for _, carFile := range carFiles {
		logs.GetLogger().Info("Uploading car file:", carFile.CarFileName)
		carFileUrl, err := client.IpfsUploadCarFileByApi(uploadUrl, carFile.CarFilePath)
		if err != nil {
			err := fmt.Errorf("failed to upload file to ipfs")
			logs.GetLogger().Error(err)
			return err
		}

		carFile.CarFileUrl = *carFileUrl //"http://" + gatewayIp + ":" + gatewayPort + "/ipfs/" + *carFileHash
		logs.GetLogger().Info("Car file: ", carFile.CarFileName, " uploaded to: ", carFile.CarFileUrl)
	}

	err = WriteCarFilesToFiles(carFiles, inputDir, constants.JSON_FILE_NAME_BY_UPLOAD, constants.CSV_FILE_NAME_BY_UPLOAD)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}
