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
