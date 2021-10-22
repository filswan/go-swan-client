package client

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go-swan-client/logs"

	ipfsApi "github.com/ipfs/go-ipfs-api"
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

func IpfsUploadCarFileByApi(uploadStreamUrl, carFilePath string) (*string, error) {
	sh := ipfsApi.NewShell(uploadStreamUrl)

	carFile, err := os.Open(carFilePath)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	defer carFile.Close()

	var carFileReader io.Reader = carFile

	cid, err := sh.Add(carFileReader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		return nil, err
	}

	carFileUrl := filepath.Join(uploadStreamUrl, "ipfs", cid)

	return &carFileUrl, nil
}
