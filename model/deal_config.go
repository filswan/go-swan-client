package model

import "github.com/shopspring/decimal"

type DealConfig struct {
	MinerFid         string
	SenderWallet     string
	MaxPrice         decimal.Decimal
	VerifiedDeal     bool
	FastRetrieval    bool
	SkipConfirmation bool
	MinerPrice       decimal.Decimal
	StartEpoch       int
}
