package client

import (
	"strings"

	"go-swan-client/logs"
)

func IpfsUploadCarFile(carFilePath string) *string {
	cmd := "ipfs add " + carFilePath
	logs.GetLogger().Info(cmd)

	result, err := ExecOsCmd(cmd, false)

	errMsg := "Failed to upload file to ipfs server."
	if err != nil {
		logs.GetLogger().Error(err)
		logs.GetLogger().Fatal(errMsg)
	}

	if result == "" {
		logs.GetLogger().Fatal(errMsg)
	}

	words := strings.Fields(result)
	if len(words) < 2 {
		logs.GetLogger().Fatal(errMsg)
	}

	carFileHash := words[1]

	return &carFileHash
}
