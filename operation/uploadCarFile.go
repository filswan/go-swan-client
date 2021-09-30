package operation

import (
	"encoding/json"
	"go-swan-client/common/utils"
	"go-swan-client/config"
	"go-swan-client/logs"
	"io/ioutil"
	"strings"
)

func UploadCarFiles(inputDir string) {
	storageServerType := config.GetConfig().Main.StorageServerType
	if storageServerType == "web server" {
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

	jsonFilePath := utils.GetDir(inputDir, "car.json")
	contents, err := ioutil.ReadFile(jsonFilePath)
	if err != nil {
		logs.GetLogger().Error("Failed to read: ", inputDir)
		return
	}

	carFiles := []*FileDesc{}

	err = json.Unmarshal(contents, &carFiles)
	if err != nil {
		logs.GetLogger().Error("Failed to read: ", inputDir)
		return
	}

	for _, carFile := range carFiles {
		logs.GetLogger().Info("Uploading car file:", carFile.CarFileName)
		car_file_hash = SwanClient.upload_car_to_ipfs(car_file.car_file_path)
		carFile.CarFileAddress = "http://" + gatewayIp + ":" + gatewayPort + "/ipfs/" + car_file_hash
		logs.GetLogger().Info("Car file: ", carFile.CarFileName, " uploaded to: ", carFile.CarFileAddress)
	}

}
