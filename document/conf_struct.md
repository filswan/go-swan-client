# Groups

* [ConfCar](#ConfCar)
* [ConfDeal](#ConfDeal)
* [ConfTask](#ConfTask)
* [ConfUpload](#ConfUpload)

## ConfCar
```shell
type ConfCar struct {
	LotusClientApiUrl      string
	LotusClientAccessToken string
	OutputDir              string
	InputDir               string
	GocarFileSizeLimit     int64
	GenerateMd5            bool
}
```

## ConfDeal
```shell
type ConfDeal struct {
	SwanApiUrl              string
	SwanApiKey              string
	SwanAccessToken         string
	SwanJwtToken            string
	LotusClientApiUrl       string
	LotusClientAccessToken  string
	SenderWallet            string
	MaxPrice                decimal.Decimal
	VerifiedDeal            bool
	FastRetrieval           bool
	SkipConfirmation        bool
	Duration                int
	MinerPrice              decimal.Decimal
	StartEpoch              int
	StartEpochIntervalHours int
	OutputDir               string
	MinerFid                string
	MetadataJsonPath        string
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
	MaxPrice                   decimal.Decimal
	StorageServerType          string
	WebServerDownloadUrlPrefix string
	ExpireDays                 int
	GenerateMd5                bool
	Duration                   int
	OutputDir                  string
	InputDir                   string
	TaskName                   string
	MinerFid                   string
	Dataset                    string
	Description                string
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
