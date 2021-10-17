package subcommand

import (
	"strings"

	"go-swan-client/common/client"
	"go-swan-client/common/constants"
	"go-swan-client/config"
	"go-swan-client/logs"
)

func UploadCarFiles(inputDir string) {
	storageServerType := config.GetConfig().Main.StorageServerType
	if storageServerType == constants.STORAGE_SERVER_TYPE_WEB_SERVER {
		logs.GetLogger().Info("Please upload car files to web server manually.")
		return
	}

	gatewayAddress := config.GetConfig().IpfsServer.GatewayAddress
	words := strings.Split(gatewayAddress, "/")
	if len(words) < 5 {
		logs.GetLogger().Fatal("Invalid gateway address:", gatewayAddress)
	}
	gatewayIp := words[2]
	gatewayPort := words[4]

	carFiles := ReadCarFilesFromJsonFile(inputDir, constants.JSON_FILE_NAME_BY_CAR)
	if carFiles == nil {
		logs.GetLogger().Fatal("Failed to read: ", inputDir)
	}

	for _, carFile := range carFiles {
		logs.GetLogger().Info("Uploading car file:", carFile.CarFileName)
		carFileHash := client.IpfsUploadCarFile(carFile.CarFilePath)
		if carFileHash == nil {
			logs.GetLogger().Fatal("Failed to upload file to ipfs.")
		}

		carFile.CarFileUrl = "http://" + gatewayIp + ":" + gatewayPort + "/ipfs/" + *carFileHash
		logs.GetLogger().Info("Car file: ", carFile.CarFileName, " uploaded to: ", carFile.CarFileUrl)
	}

	WriteCarFilesToFiles(carFiles, inputDir, constants.JSON_FILE_NAME_BY_UPLOAD, constants.CSV_FILE_NAME_BY_UPLOAD)
}
