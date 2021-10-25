package model

import (
	"fmt"
	"go-swan-client/common/constants"
	"go-swan-client/common/utils"
	"go-swan-client/config"
	"go-swan-client/logs"
	"path/filepath"
	"time"

	"github.com/shopspring/decimal"
)

type ConfDeal struct {
	SwanApiUrl              string
	SwanApiKey              string
	SwanAccessToken         string
	SenderWallet            string
	MaxPrice                decimal.Decimal
	VerifiedDeal            bool
	FastRetrieval           bool
	SkipConfirmation        bool
	MinerPrice              decimal.Decimal
	StartEpoch              int
	StartEpochIntervalHours int
	OutputDir               string
	MinerFid                *string
	MetadataJsonPath        *string
}

func GetConfDeal(outputDir *string, minerFid *string, metadataJsonPath *string) *ConfDeal {
	startEpochIntervalHours := config.GetConfig().Sender.StartEpochHours
	startEpoch := utils.GetCurrentEpoch() + (startEpochIntervalHours+1)*constants.EPOCH_PER_HOUR

	confDeal := &ConfDeal{
		SwanApiUrl:              config.GetConfig().Main.SwanApiUrl,
		SwanApiKey:              config.GetConfig().Main.SwanApiKey,
		SwanAccessToken:         config.GetConfig().Main.SwanAccessToken,
		SenderWallet:            config.GetConfig().Sender.Wallet,
		VerifiedDeal:            config.GetConfig().Sender.VerifiedDeal,
		FastRetrieval:           config.GetConfig().Sender.FastRetrieval,
		SkipConfirmation:        config.GetConfig().Sender.SkipConfirmation,
		StartEpochIntervalHours: startEpochIntervalHours,
		StartEpoch:              startEpoch,
		OutputDir:               filepath.Join(config.GetConfig().Sender.OutputDir, time.Now().Format("2006-01-02_15:04:05")),
		MinerFid:                minerFid,
		MetadataJsonPath:        metadataJsonPath,
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

func SetDealConfig4Autobid(confDeal *ConfDeal, task Task, deal OfflineDeal) error {
	confDeal.StartEpoch = utils.GetCurrentEpoch() + (confDeal.StartEpochIntervalHours+1)*constants.EPOCH_PER_HOUR

	if task.MinerFid == nil {
		err := fmt.Errorf("no miner allocated to task")
		logs.GetLogger().Error(err)
		return err
	}
	confDeal.MinerFid = task.MinerFid

	if task.Type == nil {
		err := fmt.Errorf("task type missing")
		logs.GetLogger().Error(err)
		return err
	}
	confDeal.VerifiedDeal = *task.Type == constants.TASK_TYPE_VERIFIED

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

	return nil
}
