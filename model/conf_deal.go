package model

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/filswan/go-swan-client/config"

	"github.com/filswan/go-swan-lib/constants"
	"github.com/filswan/go-swan-lib/logs"

	"github.com/shopspring/decimal"
)

type ConfDeal struct {
	SwanApiUrlToken        string          //required
	SwanApiUrl             string          //required
	SwanApiKey             string          //required when SwanJwtToken is not provided
	SwanAccessToken        string          //required when SwanJwtToken is not provided
	SwanToken              string          //required when SwanApiKey and SwanAccessToken are not provided
	LotusClientApiUrl      string          //required
	LotusClientAccessToken string          //required
	SenderWallet           string          //required
	MaxPrice               decimal.Decimal //required only for manual-bid deal
	VerifiedDeal           bool            //required only for manual-bid deal
	FastRetrieval          bool            //required only for manual-bid deal
	SkipConfirmation       bool            //required only for manual-bid deal
	Duration               int             //not necessary, when not provided use default value:1512000
	StartEpochHours        int             //required only for manual-bid deal
	OutputDir              string          //required
	MinerFids              []string        //required only for manual-bid deal
	MetadataJsonPath       string          //required only for manual-bid deal
	DealSourceIds          []int           //required
}

func GetConfDeal(outputDir *string, minerFids string, metadataJsonPath string) *ConfDeal {
	confDeal := &ConfDeal{
		SwanApiUrlToken:        config.GetConfig().Main.SwanApiUrlToken,
		SwanApiUrl:             config.GetConfig().Main.SwanApiUrl,
		SwanApiKey:             config.GetConfig().Main.SwanApiKey,
		SwanAccessToken:        config.GetConfig().Main.SwanAccessToken,
		LotusClientApiUrl:      config.GetConfig().Lotus.ClientApiUrl,
		LotusClientAccessToken: config.GetConfig().Lotus.ClientAccessToken,
		SenderWallet:           config.GetConfig().Sender.Wallet,
		VerifiedDeal:           config.GetConfig().Sender.VerifiedDeal,
		FastRetrieval:          config.GetConfig().Sender.FastRetrieval,
		SkipConfirmation:       config.GetConfig().Sender.SkipConfirmation,
		Duration:               config.GetConfig().Sender.Duration,
		StartEpochHours:        config.GetConfig().Sender.StartEpochHours,
		OutputDir:              filepath.Join(config.GetConfig().Sender.OutputDir, time.Now().Format("2006-01-02_15:04:05")),
		MinerFids:              strings.Split(minerFids, ","),
		MetadataJsonPath:       metadataJsonPath,
	}

	confDeal.DealSourceIds = append(confDeal.DealSourceIds, constants.TASK_SOURCE_ID_SWAN)
	confDeal.DealSourceIds = append(confDeal.DealSourceIds, constants.TASK_SOURCE_ID_SWAN_CLIENT)

	if outputDir != nil && len(*outputDir) != 0 {
		confDeal.OutputDir = *outputDir
	}

	maxPriceStr := config.GetConfig().Sender.MaxPrice
	maxPrice, err := decimal.NewFromString(maxPriceStr)
	if err != nil {
		logs.GetLogger().Error("Failed to convert maxPrice(" + maxPriceStr + ") to decimal, MaxPrice:")
		return nil
	}
	confDeal.MaxPrice = maxPrice

	return confDeal
}
