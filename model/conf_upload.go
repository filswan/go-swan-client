package model

import (
	"go-swan-client/config"
	"time"
)

type ConfUpload struct {
	StorageServerType           string
	IpfsServerDownloadUrlPrefix string
	OutputDir                   string
}

func GetConfUpload() *ConfUpload {
	confUpload := &ConfUpload{
		StorageServerType:           config.GetConfig().Main.StorageServerType,
		IpfsServerDownloadUrlPrefix: config.GetConfig().IpfsServer.DownloadUrlPrefix,
		OutputDir:                   config.GetConfig().Sender.OutputDir + time.Now().Format("2006-01-02_15:04:05"),
	}

	return confUpload
}
