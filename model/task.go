package model

type Task struct {
	TaskName       string  `json:"task_name"`
	CuratedDataset string  `json:"curated_dataset"`
	Description    string  `json:"description"`
	IsPublic       bool    `json:"is_public"`
	TaskType       string  `json:"type"`
	MinerId        *string `json:"miner_id"`
	FastRetrieval  bool    `fast_retrieval"`
	BidMode        int     `bid_mode"`
	MaxPrice       string  `max_price"`
	ExpireDays     int     `expire_days"`
	Uuid           string  `json:"uuid"`
	//IsVerified     bool    `json:"is_verified"`
}
