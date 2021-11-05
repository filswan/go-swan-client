package model

import (
	"path/filepath"
	"time"

	"github.com/filswan/go-swan-client/config"

	"github.com/filswan/go-swan-lib/constants"
	"github.com/filswan/go-swan-lib/logs"
	"github.com/filswan/go-swan-lib/utils"
	"github.com/shopspring/decimal"
)

type ConfTask struct {
	SwanApiUrl                 string
	SwanApiKey                 string
	SwanAccessToken            string
	SwanJwtToken               string
	PublicDeal                 bool
	BidMode                    int
	VerifiedDeal               bool
	OfflineMode                bool
	FastRetrieval              bool
	MaxPrice                   decimal.Decimal
	StorageServerType          string
	WebServerDownloadUrlPrefix string
	ExpireDays                 int
	GenerateMd5                bool
	Duration                   int
	OutputDir                  string
	InputDir                   string
	TaskName                   string
	MinerFid                   string
	Dataset                    string
	Description                string
	StartEpoch                 int
	StartEpochIntervalHours    int
	SourceId                   int
}

func GetConfTask(inputDir string, outputDir *string, taskName, minerFid, dataset, description string) *ConfTask {
	startEpochIntervalHours := config.GetConfig().Sender.StartEpochHours
	startEpoch := utils.GetCurrentEpoch() + (startEpochIntervalHours+1)*constants.EPOCH_PER_HOUR

	confTask := &ConfTask{
		SwanApiUrl:                 config.GetConfig().Main.SwanApiUrl,
		SwanApiKey:                 config.GetConfig().Main.SwanApiKey,
		SwanAccessToken:            config.GetConfig().Main.SwanAccessToken,
		PublicDeal:                 config.GetConfig().Sender.PublicDeal,
		BidMode:                    config.GetConfig().Sender.BidMode,
		VerifiedDeal:               config.GetConfig().Sender.VerifiedDeal,
		OfflineMode:                config.GetConfig().Sender.OfflineMode,
		FastRetrieval:              config.GetConfig().Sender.FastRetrieval,
		StorageServerType:          config.GetConfig().Main.StorageServerType,
		WebServerDownloadUrlPrefix: config.GetConfig().WebServer.DownloadUrlPrefix,
		ExpireDays:                 config.GetConfig().Sender.ExpireDays,
		GenerateMd5:                config.GetConfig().Sender.GenerateMd5,
		Duration:                   config.GetConfig().Sender.Duration,
		OutputDir:                  filepath.Join(config.GetConfig().Sender.OutputDir, time.Now().Format("2006-01-02_15:04:05")),
		InputDir:                   inputDir,
		TaskName:                   taskName,
		MinerFid:                   minerFid,
		Dataset:                    dataset,
		Description:                description,
		StartEpochIntervalHours:    startEpochIntervalHours,
		StartEpoch:                 startEpoch,
		SourceId:                   constants.TASK_SOURCE_ID_SWAN_CLIENT,
	}

	if outputDir != nil && len(*outputDir) != 0 {
		confTask.OutputDir = *outputDir
	}

	var err error
	maxPrice := config.GetConfig().Sender.MaxPrice
	confTask.MaxPrice, err = decimal.NewFromString(maxPrice)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	return confTask
}
