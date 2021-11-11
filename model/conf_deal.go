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
	SwanApiUrl                   string
	SwanApiKey                   string
	SwanAccessToken              string
	SwanJwtToken                 string
	LotusClientApiUrl            string
	LotusClientAccessToken       string
	SenderWallet                 string
	MaxPrice                     decimal.Decimal
	VerifiedDeal                 bool
	FastRetrieval                bool
	SkipConfirmation             bool
	Duration                     int
	MinerPrice                   decimal.Decimal
	StartEpoch                   int
	StartEpochIntervalHours      int
	OutputDir                    string
	MinerFid                     string
	MetadataJsonPath             string
	DealSourceIds                []int
	RelativeEpochFromMainNetwork int
}

func GetConfDeal(outputDir *string, minerFid, metadataJsonPath string, isAutoBid bool) *ConfDeal {
	startEpochIntervalHours := config.GetConfig().Sender.StartEpochHours
	startEpoch := utils.GetCurrentEpoch() + (startEpochIntervalHours+1)*constants.EPOCH_PER_HOUR
	startEpoch = startEpoch + config.GetConfig().Sender.RelativeEpochFromMainNetwork

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

	if isAutoBid {
		confDeal.SkipConfirmation = true
	}

	if outputDir != nil && len(*outputDir) != 0 {
		confDeal.OutputDir = *outputDir
	}

	logs.GetLogger().Info(confDeal.OutputDir)

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

	return nil
}
