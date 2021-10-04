package subcommand

import (
	"go-swan-client/common/constants"
	"go-swan-client/common/utils"
	"go-swan-client/config"
	"go-swan-client/logs"
	"strings"
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
		logs.GetLogger().Error("Invalid gateway address:", gatewayAddress)
	}
	gatewayIp := words[2]
	gatewayPort := words[4]

	carFiles := readCarFilesFromJsonFile(inputDir)
	if carFiles == nil {
		logs.GetLogger().Error("Failed to read: ", inputDir)
		return
	}

	for _, carFile := range carFiles {
		logs.GetLogger().Info("Uploading car file:", carFile.CarFileName)
		carFileHash := utils.IpfsUploadCarFile(carFile.CarFilePath)
		if carFileHash == nil {
			logs.GetLogger().Error("Failed to upload file to ipfs.")
			return
		}

		carFile.CarFileUrl = "http://" + gatewayIp + ":" + gatewayPort + "/ipfs/" + *carFileHash
		logs.GetLogger().Info("Car file: ", carFile.CarFileName, " uploaded to: ", carFile.CarFileUrl)
	}

	generateCsvFile(carFiles, inputDir, "car.csv")
	generateJsonFile(carFiles, inputDir)
}
