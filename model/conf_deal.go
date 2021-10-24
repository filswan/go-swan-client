package model

import "github.com/shopspring/decimal"

type ConfDeal struct {
	MinerFid         string
	SenderWallet     string
	MaxPrice         decimal.Decimal
	VerifiedDeal     bool
	FastRetrieval    bool
	SkipConfirmation bool
	MinerPrice       decimal.Decimal
	StartEpoch       int
}
