package model

import (
	"github.com/filswan/go-swan-client/config"
)

type ConfUpload struct {
	StorageServerType           string //required
	IpfsServerDownloadUrlPrefix string //required only when upload to ipfs server
	IpfsServerUploadUrlPrefix   string //required only when upload to ipfs server
	OutputDir                   string //invalid
	InputDir                    string //required
}

func GetConfUpload(inputDir string) *ConfUpload {
	confUpload := &ConfUpload{
		StorageServerType:           config.GetConfig().Main.StorageServerType,
		IpfsServerDownloadUrlPrefix: config.GetConfig().IpfsServer.DownloadUrlPrefix,
		IpfsServerUploadUrlPrefix:   config.GetConfig().IpfsServer.UploadUrlPrefix,
		OutputDir:                   inputDir,
		InputDir:                    inputDir,
	}

	return confUpload
}
