package model

import (
	"go-swan-client/common/constants"
	"go-swan-client/common/utils"
	"go-swan-client/config"
	"go-swan-client/logs"

	"github.com/shopspring/decimal"
)

type ConfDeal struct {
	SwanApiUrl       string
	SwanApiKey       string
	SwanAccessToken  string
	MinerFid         string
	SenderWallet     string
	MaxPrice         decimal.Decimal
	VerifiedDeal     bool
	FastRetrieval    bool
	SkipConfirmation bool
	MinerPrice       decimal.Decimal
	StartEpoch       int
}

func GetConfDeal(minerFid *string) *ConfDeal {
	startEpochIntervalHours := config.GetConfig().Sender.StartEpochHours + 1
	startEpoch := utils.GetCurrentEpoch() + startEpochIntervalHours*constants.EPOCH_PER_HOUR

	dealConfig := ConfDeal{
		SwanApiUrl:       config.GetConfig().Main.SwanApiUrl,
		SwanApiKey:       config.GetConfig().Main.SwanApiKey,
		SwanAccessToken:  config.GetConfig().Main.SwanAccessToken,
		SenderWallet:     config.GetConfig().Sender.Wallet,
		VerifiedDeal:     config.GetConfig().Sender.VerifiedDeal,
		FastRetrieval:    config.GetConfig().Sender.FastRetrieval,
		SkipConfirmation: config.GetConfig().Sender.SkipConfirmation,
		StartEpoch:       startEpoch,
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

func GetDealConfig4Autobid(task Task, deal OfflineDeal) *ConfDeal {
	startEpochIntervalHours := config.GetConfig().Sender.StartEpochHours + 1
	startEpoch := utils.GetCurrentEpoch() + startEpochIntervalHours*constants.EPOCH_PER_HOUR

	dealConfig := ConfDeal{
		MinerFid:         *task.MinerFid,
		SenderWallet:     config.GetConfig().Sender.Wallet,
		VerifiedDeal:     *task.Type == constants.TASK_TYPE_VERIFIED,
		FastRetrieval:    *task.FastRetrieval == constants.TASK_FAST_RETRIEVAL,
		SkipConfirmation: config.GetConfig().Sender.SkipConfirmation,
		StartEpoch:       startEpoch,
	}

	dealConfig.MaxPrice = *task.MaxPrice

	return &dealConfig
}
