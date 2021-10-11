package model

type DealConfig struct {
	MinerFid           string
	SenderWallet       string
	MaxPrice           float64
	VerifiedDeal       bool
	FastRetrieval      bool
	EpochIntervalHours int
	SkipConfirmation   bool
	MinerPrice         float64
	StartEpochHours    int
}
