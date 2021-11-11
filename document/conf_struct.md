# Groups

* [ConfCar](#ConfCar)
* [ConfUpload](#ConfUpload)
* [ConfTask](#ConfTask)
* [ConfDeal](#ConfDeal)

## ConfCar
```shell
type ConfCar struct {
	LotusClientApiUrl      string //required
	LotusClientAccessToken string //required
	OutputDir              string //required
	InputDir               string //required
	GocarFileSizeLimit     int64  //required only when creating gocar file(s)
	GenerateMd5            bool   //required
}
```

## ConfUpload
```shell
type ConfUpload struct {
	StorageServerType           string //required
	IpfsServerDownloadUrlPrefix string //required only when upload to ipfs server
	IpfsServerUploadUrl         string //required only when upload to ipfs server
	OutputDir                   string //invalid
	InputDir                    string //required
}
```

## ConfTask
```shell
type ConfTask struct {
	SwanApiUrl                 string          //required when OfflineMode is false
	SwanApiKey                 string          //required when OfflineMode is false and SwanJwtToken is not provided
	SwanAccessToken            string          //required when OfflineMode is false and SwanJwtToken is not provided
	SwanJwtToken               string          //required when OfflineMode is false and SwanApiKey & SwanAccessToken are not provided
	PublicDeal                 bool            //required
	BidMode                    int             //required
	VerifiedDeal               bool            //required
	OfflineMode                bool            //required
	FastRetrieval              bool            //required
	MaxPrice                   decimal.Decimal //required
	StorageServerType          string          //required
	WebServerDownloadUrlPrefix string          //required only when StorageServerType is web server
	ExpireDays                 int             //required
	GenerateMd5                bool            //required
	Duration                   int             //not necessary, when not provided use default value:1512000
	OutputDir                  string          //required
	InputDir                   string          //required
	TaskName                   string          //not necessary, when not provided use default value:swan_task_xxxxxx
	MinerFid                   string          //required only when PublicDeal=false
	Dataset                    string          //not necessary
	Description                string          //not necessary
	StartEpoch                 int             //required
	StartEpochIntervalHours    int             //invalid
	SourceId                   int             //required
}
```

## ConfDeal
```shell
type ConfDeal struct {
	SwanApiUrl                   string
	SwanApiKey                   string
	SwanAccessToken              string
	SwanJwtToken                 string
	LotusClientApiUrl            string
	LotusClientAccessToken       string
	SenderWallet                 string
	MaxPrice                     decimal.Decimal
	VerifiedDeal                 bool
	FastRetrieval                bool
	SkipConfirmation             bool
	Duration                     int
	MinerPrice                   decimal.Decimal
	StartEpoch                   int
	StartEpochIntervalHours      int
	OutputDir                    string
	MinerFid                     string
	MetadataJsonPath             string
	DealSourceIds                []int
	RelativeEpochFromMainNetwork int
}
```
