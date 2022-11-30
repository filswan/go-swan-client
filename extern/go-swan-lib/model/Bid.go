package model

import "github.com/shopspring/decimal"

type Bid struct {
	Id           int              `json:"id"`
	Price        *decimal.Decimal `json:"price"`
	MinPieceSize *string          `json:"min_price_size"`
	Status       string           `json:"status"`
	CreatedOn    string           `json:"created_on"`
	WonOn        string           `json:"won_on"`
	ExpireDays   *int             `json:"expire_days"`
	MinerFid     string           `json:"miner_fid"`
}
