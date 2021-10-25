package model

import "github.com/shopspring/decimal"

//type Task struct {
//	TaskName       string  `json:"task_name"`
//	CuratedDataset string  `json:"curated_dataset"`
//	Description    string  `json:"description"`
//	IsPublic       bool    `json:"is_public"` //int
//	TaskType       string  `json:"type"`
//	MinerId        *string `json:"miner_id"`
//	FastRetrieval  bool    `fast_retrieval"`
//	BidMode        int     `bid_mode"`
//	MaxPrice       string  `max_price"`
//	ExpireDays     int     `expire_days"`
//	Uuid           string  `json:"uuid"`
//	//IsVerified     bool    `json:"is_verified"`
//}

type Task struct {
	Id                int              `json:"id"`
	TaskName          string           `json:"task_name"`
	Description       string           `json:"description"`
	TaskFileName      string           `json:"task_file_name"`
	CreatedOn         string           `json:"created_on"`
	UserId            int              `json:"user_id"`
	Status            string           `json:"status"`
	Tags              string           `json:"tags"`
	MinerFid          *string          `json:"miner_id"`
	Type              *string          `json:"type"`
	IsPublic          *int             `json:"is_public"`
	MinPrice          *decimal.Decimal `json:"min_price"`
	MaxPrice          *decimal.Decimal `json:"max_price"`
	ExpireDays        *int             `json:"expire_days"`
	Uuid              *string          `json:"uuid"`
	CuratedDataset    string           `json:"curated_dataset"`
	UpdatedOn         string           `json:"updated_on"`
	BidMode           *int             `json:"bid_mode"`
	FastRetrieval     *int             `json:"fast_retrieval"`
	FastRetrievalBool bool
}
