# Groups

* [ConfCar](#ConfCar)
* [ConfDeal](#ConfDeal)
* [ConfTask](#ConfTask)
* [ConfUpload](#ConfUpload)

## ConfCar
```shell
type ConfCar struct {
	LotusApiUrl        string
	LotusAccessToken   string
	OutputDir          string
	InputDir           string
	GocarFileSizeLimit int64
}
```

## ConfDeal
```shell
type ConfDeal struct {
	SwanApiUrl              string
	SwanApiKey              string
	SwanAccessToken         string
	SenderWallet            string
	MaxPrice                decimal.Decimal
	VerifiedDeal            bool
	FastRetrieval           bool
	SkipConfirmation        bool
	MinerPrice              decimal.Decimal
	StartEpoch              int
	StartEpochIntervalHours int
	OutputDir               string
	MinerFid                *string
	MetadataJsonPath        *string
}
```

## ConfTask
```shell
type ConfTask struct {
	SwanApiUrl                 string
	SwanApiKey                 string
	SwanAccessToken            string
	SwanJwtToken               string
	PublicDeal                 bool
	BidMode                    int
	VerifiedDeal               bool
	OfflineMode                bool
	FastRetrieval              bool
	MaxPrice                   string
	StorageServerType          string
	WebServerDownloadUrlPrefix string
	ExpireDays                 int
	OutputDir                  string
	InputDir                   string
	TaskName                   *string
	MinerFid                   *string
	Dataset                    *string
	Description                *string
	StartEpoch                 int
	StartEpochIntervalHours    int
	SourceId                   int
}
```

## ConfUpload
```shell
type ConfUpload struct {
	StorageServerType           string
	IpfsServerDownloadUrlPrefix string
	IpfsServerUploadUrl         string
	OutputDir                   string
	InputDir                    string
}
```
