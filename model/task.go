package model

type Task struct {
	TaskName       string  `json:"task_name"`
	CuratedDataset string  `json:"curated_dataset"`
	Description    string  `json:"description"`
	IsPublic       bool    `json:"is_public"`
	IsVerified     bool    `json:"is_verified"`
	FastRetrieval  bool    `fast_retrieval"`
	MaxPrice       string  `max_price"`
	BidMode        int     `bid_mode"`
	ExpireDays     int     `expire_days"`
	MinerId        *string `json:"miner_id"`
	Uuid           string  `json:"uuid"`
}
