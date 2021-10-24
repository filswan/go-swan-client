package model

import (
	"fmt"
	"go-swan-client/common/constants"
	"go-swan-client/common/utils"
	"go-swan-client/config"
	"go-swan-client/logs"

	"github.com/shopspring/decimal"
)

type ConfDeal struct {
	SwanApiUrl              string
	SwanApiKey              string
	SwanAccessToken         string
	MinerFid                string
	SenderWallet            string
	MaxPrice                decimal.Decimal
	VerifiedDeal            bool
	FastRetrieval           bool
	SkipConfirmation        bool
	MinerPrice              decimal.Decimal
	StartEpoch              int
	StartEpochIntervalHours int
}

func GetConfDeal(minerFid *string) *ConfDeal {
	startEpochIntervalHours := config.GetConfig().Sender.StartEpochHours
	startEpoch := utils.GetCurrentEpoch() + (startEpochIntervalHours+1)*constants.EPOCH_PER_HOUR

	dealConfig := ConfDeal{
		SwanApiUrl:              config.GetConfig().Main.SwanApiUrl,
		SwanApiKey:              config.GetConfig().Main.SwanApiKey,
		SwanAccessToken:         config.GetConfig().Main.SwanAccessToken,
		SenderWallet:            config.GetConfig().Sender.Wallet,
		VerifiedDeal:            config.GetConfig().Sender.VerifiedDeal,
		FastRetrieval:           config.GetConfig().Sender.FastRetrieval,
		SkipConfirmation:        config.GetConfig().Sender.SkipConfirmation,
		StartEpochIntervalHours: startEpochIntervalHours,
		StartEpoch:              startEpoch,
	}

	if minerFid != nil {
		dealConfig.MinerFid = *minerFid
	}

	maxPriceStr := config.GetConfig().Sender.MaxPrice
	maxPrice, err := decimal.NewFromString(maxPriceStr)
	if err != nil {
		logs.GetLogger().Error("Failed to convert maxPrice(" + maxPriceStr + ") to decimal, MaxPrice:")
		return nil
	}
	dealConfig.MaxPrice = maxPrice

	return &dealConfig
}

func SetDealConfig4Autobid(confDeal *ConfDeal, task Task, deal OfflineDeal) error {
	confDeal.StartEpoch = utils.GetCurrentEpoch() + (confDeal.StartEpochIntervalHours+1)*constants.EPOCH_PER_HOUR

	if task.MinerFid == nil {
		err := fmt.Errorf("no miner allocated to task")
		logs.GetLogger().Error(err)
		return err
	}
	confDeal.MinerFid = *task.MinerFid

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
