package model

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/filswan/go-swan-client/config"

	"github.com/filswan/go-swan-lib/constants"
	"github.com/filswan/go-swan-lib/logs"
	libmodel "github.com/filswan/go-swan-lib/model"
	"github.com/filswan/go-swan-lib/utils"

	"github.com/shopspring/decimal"
)

type ConfDeal struct {
	SwanApiUrl                   string          //required
	SwanApiKey                   string          //required when SwanJwtToken is not provided
	SwanAccessToken              string          //required when SwanJwtToken is not provided
	SwanToken                    string          //required when SwanApiKey and SwanAccessToken are not provided
	LotusClientApiUrl            string          //required
	LotusClientAccessToken       string          //required
	SenderWallet                 string          //required
	MaxPrice                     decimal.Decimal //required only for manual-bid deal
	VerifiedDeal                 bool            //required only for manual-bid deal
	FastRetrieval                bool            //required only for manual-bid deal
	SkipConfirmation             bool            //required only for manual-bid deal
	Duration                     int             //not necessary, when not provided use default value:1512000
	MinerPrice                   decimal.Decimal //used internally, not need to provide
	StartEpoch                   int             //required only for manual-bid deal
	StartEpochIntervalHours      int             //invalid
	OutputDir                    string          //required
	MinerFid                     string          //required only for manual-bid deal
	MetadataJsonPath             string          //required only for manual-bid deal
	DealSourceIds                []int           //required
	RelativeEpochFromMainNetwork int             //required
}

func GetConfDeal(outputDir *string, minerFid, metadataJsonPath string) *ConfDeal {
	startEpochIntervalHours := config.GetConfig().Sender.StartEpochHours
	startEpoch := utils.GetCurrentEpoch()
	startEpoch = startEpoch + config.GetConfig().Sender.RelativeEpochFromMainNetwork
	logs.GetLogger().Info("current epoch:", startEpoch)
	startEpoch = startEpoch + (startEpochIntervalHours+1)*constants.EPOCH_PER_HOUR
	logs.GetLogger().Info("start epoch:", startEpoch)

	confDeal := &ConfDeal{
		SwanApiUrl:                   config.GetConfig().Main.SwanApiUrl,
		SwanApiKey:                   config.GetConfig().Main.SwanApiKey,
		SwanAccessToken:              config.GetConfig().Main.SwanAccessToken,
		LotusClientApiUrl:            config.GetConfig().Lotus.ClientApiUrl,
		LotusClientAccessToken:       config.GetConfig().Lotus.ClientAccessToken,
		SenderWallet:                 config.GetConfig().Sender.Wallet,
		VerifiedDeal:                 config.GetConfig().Sender.VerifiedDeal,
		FastRetrieval:                config.GetConfig().Sender.FastRetrieval,
		SkipConfirmation:             config.GetConfig().Sender.SkipConfirmation,
		Duration:                     config.GetConfig().Sender.Duration,
		StartEpochIntervalHours:      startEpochIntervalHours,
		StartEpoch:                   startEpoch,
		OutputDir:                    filepath.Join(config.GetConfig().Sender.OutputDir, time.Now().Format("2006-01-02_15:04:05")),
		MinerFid:                     minerFid,
		MetadataJsonPath:             metadataJsonPath,
		RelativeEpochFromMainNetwork: config.GetConfig().Sender.RelativeEpochFromMainNetwork,
	}

	confDeal.DealSourceIds = append(confDeal.DealSourceIds, constants.TASK_SOURCE_ID_SWAN)
	confDeal.DealSourceIds = append(confDeal.DealSourceIds, constants.TASK_SOURCE_ID_SWAN_CLIENT)

	if outputDir != nil && len(*outputDir) != 0 {
		confDeal.OutputDir = *outputDir
	}

	//logs.GetLogger().Info(confDeal.OutputDir)

	maxPriceStr := config.GetConfig().Sender.MaxPrice
	maxPrice, err := decimal.NewFromString(maxPriceStr)
	if err != nil {
		logs.GetLogger().Error("Failed to convert maxPrice(" + maxPriceStr + ") to decimal, MaxPrice:")
		return nil
	}
	confDeal.MaxPrice = maxPrice

	return confDeal
}

func SetDealConfig4Autobid(confDeal *ConfDeal, task libmodel.Task, deal libmodel.OfflineDeal) error {
	if confDeal == nil {
		err := fmt.Errorf("parameter confDeal is nil")
		logs.GetLogger().Error(err)
		return err
	}

	confDeal.StartEpoch = deal.StartEpoch + confDeal.RelativeEpochFromMainNetwork

	if task.MinerFid == "" {
		err := fmt.Errorf("no miner allocated to task")
		logs.GetLogger().Error(err)
		return err
	}
	confDeal.MinerFid = task.MinerFid

	if task.Type == "" {
		err := fmt.Errorf("task type missing")
		logs.GetLogger().Error(err)
		return err
	}
	confDeal.VerifiedDeal = task.Type == constants.TASK_TYPE_VERIFIED

	if task.FastRetrieval == nil {
		err := fmt.Errorf("task FastRetrieval missing")
		logs.GetLogger().Error(err)
		return err
	}

	confDeal.FastRetrieval = *task.FastRetrieval == constants.TASK_FAST_RETRIEVAL

	if task.MaxPrice == nil {
		err := fmt.Errorf("task MaxPrice missing")
		logs.GetLogger().Error(err)
		return err
	}
	confDeal.MaxPrice = *task.MaxPrice

	confDeal.Duration = task.Duration

	confDeal.SkipConfirmation = true

	return nil
}
