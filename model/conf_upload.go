package model

import "go-swan-client/config"

type ConfUpload struct {
	StorageServerType           string
	IpfsServerDownloadUrlPrefix string
	OutputDir                   string
}

func GetConfUpload() *ConfUpload {
	confUpload := &ConfUpload{
		StorageServerType:           config.GetConfig().Main.StorageServerType,
		IpfsServerDownloadUrlPrefix: config.GetConfig().IpfsServer.DownloadUrlPrefix,
		OutputDir:                   config.GetConfig().Sender.OutputDir,
	}

	return confUpload
}
