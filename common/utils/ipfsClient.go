package utils

import (
	"go-swan-client/logs"
	"strings"
)

func IpfsUploadCarFile(carFilePath string) *string {
	cmd := "ipfs add " + carFilePath + " | grep added"
	logs.GetLogger().Info(cmd)

	result, err := ExecOsCmd(cmd)

	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	if result == "" {
		logs.GetLogger().Error("Failed to upload file to ipfs server.")
		return nil
	}

	words := strings.Split(result, " ")
	if len(words) < 2 {
		logs.GetLogger().Error("Failed to upload file to ipfs server.")
		return nil
	}

	carFileHash := words[1]

	return &carFileHash
}
