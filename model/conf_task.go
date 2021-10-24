package model

import (
	"go-swan-client/config"
	"time"
)

type ConfTask struct {
	SwanApiUrl                 string
	SwanApiKey                 string
	SwanAccessToken            string
	PublicDeal                 bool
	BidMode                    int
	VerifiedDeal               bool
	OfflineMode                bool
	FastRetrieval              bool
	MaxPrice                   string
	StorageServerType          string
	WebServerDownloadUrlPrefix string
	ExpireDays                 int
	OutputDir                  string
}

func GetConfTask(outDir *string) *ConfTask {
	confTask := &ConfTask{
		SwanApiUrl:                 config.GetConfig().Main.SwanApiUrl,
		SwanApiKey:                 config.GetConfig().Main.SwanApiKey,
		SwanAccessToken:            config.GetConfig().Main.SwanAccessToken,
		PublicDeal:                 config.GetConfig().Sender.PublicDeal,
		BidMode:                    config.GetConfig().Sender.BidMode,
		VerifiedDeal:               config.GetConfig().Sender.VerifiedDeal,
		OfflineMode:                config.GetConfig().Sender.OfflineMode,
		FastRetrieval:              config.GetConfig().Sender.FastRetrieval,
		MaxPrice:                   config.GetConfig().Sender.MaxPrice,
		StorageServerType:          config.GetConfig().Main.StorageServerType,
		WebServerDownloadUrlPrefix: config.GetConfig().WebServer.DownloadUrlPrefix,
		ExpireDays:                 config.GetConfig().Sender.ExpireDays,
		OutputDir:                  config.GetConfig().Sender.OutputDir + time.Now().Format("2006-01-02_15:04:05"),
	}

	if outDir != nil {
		confTask.OutputDir = *outDir
	}

	return confTask
}
