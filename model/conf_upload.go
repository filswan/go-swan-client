package model

import (
	"go-swan-client/config"
)

type ConfUpload struct {
	StorageServerType           string
	IpfsServerDownloadUrlPrefix string
	OutputDir                   string
	InputDir                    string
}

func GetConfUpload(inputDir string) *ConfUpload {
	confUpload := &ConfUpload{
		StorageServerType:           config.GetConfig().Main.StorageServerType,
		IpfsServerDownloadUrlPrefix: config.GetConfig().IpfsServer.DownloadUrlPrefix,
		OutputDir:                   inputDir,
		InputDir:                    inputDir,
	}

	return confUpload
}
