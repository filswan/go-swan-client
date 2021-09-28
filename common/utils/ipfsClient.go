package utils

import (
	"go-swan-client/logs"
	"strings"
)

func UploadCar2Ipfs(carFilePath string) *string {
	cmd := "ipfs add " + carFilePath + " | grep added"
	logs.GetLogger().Info(cmd)

	result, err := ExecOsCmd(cmd)

	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	if result == "" {
		logs.GetLogger().Error("Upload file to ipfs server failed.")
		return nil
	}

	carFileHash := strings.Split(result, " ")[1]

	return &carFileHash
}
