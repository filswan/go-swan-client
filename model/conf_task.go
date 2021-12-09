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
	SwanApiUrl                 string          //required when OfflineMode is false
	SwanApiKey                 string          //required when OfflineMode is false and SwanJwtToken is not provided
	SwanAccessToken            string          //required when OfflineMode is false and SwanJwtToken is not provided
	SwanToken                  string          //required when OfflineMode is false and SwanApiKey & SwanAccessToken are not provided
	PublicDeal                 bool            //required
	BidMode                    int             //required
	VerifiedDeal               bool            //required
	OfflineMode                bool            //required
	FastRetrieval              bool            //required
	MaxPrice                   decimal.Decimal //required
	StorageServerType          string          //required
	WebServerDownloadUrlPrefix string          //required only when StorageServerType is web server
	ExpireDays                 int             //required
	GenerateMd5                bool            //required
	Duration                   int             //not necessary, when not provided use default value:1512000
	OutputDir                  string          //required
	InputDir                   string          //required
	TaskName                   string          //not necessary, when not provided use default value:swan_task_xxxxxx
	Dataset                    string          //not necessary
	Description                string          //not necessary
	StartEpoch                 int             //required
	StartEpochIntervalHours    int             //invalid
	SourceId                   int             //required
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
