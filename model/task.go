package model

type Task struct {
	TaskName       string  `json:"task_name"`
	CuratedDataset string  `json:"curated_dataset"`
	Description    string  `json:"description"`
	IsPublic       bool    `json:"is_public"`
	IsVerified     bool    `json:"is_verified"`
	MinerId        *string `json:"miner_id"`
	Uuid           string  `json:"uuid`
}
