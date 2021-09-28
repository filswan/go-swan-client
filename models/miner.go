package models

type Miner struct {
	Id                     int      `json:"id"`
	Score                  int      `json:"score"`
	Price                  *float64 `json:"price"`
	VerifiedPrice          *float64 `json:"verified_price"`
	MinerFid               string   `json:"miner_fid"`
	UpdateTimeStr          string   `json:"update_time_str"`
	Status                 string   `json:"status"`
	MinPieceSize           *string  `json:"min_piece_size"`
	MaxPieceSize           *string  `json:"max_piece_size"`
	Location               string   `json:"location"`
	SwanMinerId            int      `json:"swan_miner_id"`
	OfflineDealAvailable   int      `json:"offline_deal_available"`
	DailySealingCapability int      `json:"daily_sealing_capabilty"`
	AdjustedPower          string   `json:"adjusted_power"`
	ReachableCount         int      `json:"reachable_count"`
	UnreachableCount       int      `json:"unreachable_count"`
	DailyReward            float64  `json:"daily_reward"`
	SectorLiveCount        int      `json:"sector_live_count"`
	SectorFaultyCount      int      `json:"sector_faulty_count"`
	SectorActiveCount      int      `json:"sector_active_count"`
	BidMode                int      `json:"bid_mode"`
	StartEpoch             *int     `json:"start_epoch"`
	AddressBalance         float64  `json:"address_balance"`
	AutoBidTaskPerDay      int      `json:"auto_bid_task_per_day"`
	AutoBidTaskCnt         int      `json:"auto_bid_task_cnt"`
	LastAutoBidAt          int64    `json:"last_auto_bid_at"` //millisecond of last auto-bid task for this miner
	ExpectedSealingTime    *int     `json:"expected_sealing_time"`
}
