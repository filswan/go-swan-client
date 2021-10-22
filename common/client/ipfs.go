package client

import (
	"fmt"
	"strings"

	"go-swan-client/logs"
)

func IpfsUploadCarFile(carFilePath string) (*string, error) {
	cmd := "ipfs add " + carFilePath
	logs.GetLogger().Info(cmd)

	result, err := ExecOsCmd(cmd, false)

	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if result == "" {
		err := fmt.Errorf("cmd(%s) result is empty", cmd)
		logs.GetLogger().Error(err)
		return nil, err
	}

	words := strings.Fields(result)
	if len(words) < 2 {
		err := fmt.Errorf("cmd(%s) result(%s) does not have enough fields", cmd, result)
		logs.GetLogger().Error(err)
		return nil, err
	}

	carFileHash := words[1]

	return &carFileHash, nil
}
