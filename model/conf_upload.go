package model

import (
	"github.com/filswan/go-swan-client/config"
)

type ConfUpload struct {
	StorageServerType           string
	IpfsServerDownloadUrlPrefix string
	IpfsServerUploadUrl         string
	OutputDir                   string
	InputDir                    string
}

func GetConfUpload(inputDir string) *ConfUpload {
	confUpload := &ConfUpload{
		StorageServerType:           config.GetConfig().Main.StorageServerType,
		IpfsServerDownloadUrlPrefix: config.GetConfig().IpfsServer.DownloadUrlPrefix,
		IpfsServerUploadUrl:         config.GetConfig().IpfsServer.UploadUrl,
		OutputDir:                   inputDir,
		InputDir:                    inputDir,
	}

	return confUpload
}
