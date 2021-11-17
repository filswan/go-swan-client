# Groups

* [ConfCar](#ConfCar)
* [ConfUpload](#ConfUpload)
* [ConfTask](#ConfTask)
* [ConfDeal](#ConfDeal)
* [Deal](#Deal)

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
	SwanToken                  string          //required when OfflineMode is false and SwanApiKey & SwanAccessToken are not provided
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
	SwanApiUrl                   string          //required
	SwanApiKey                   string          //required when SwanJwtToken is not provided
	SwanAccessToken              string          //required when SwanJwtToken is not provided
	SwanToken                    string          //required when SwanApiKey and SwanAccessToken are not provided
	LotusClientApiUrl            string          //required
	LotusClientAccessToken       string          //required
	SenderWallet                 string          //required
	MaxPrice                     decimal.Decimal //required only for manual-bid deal
	VerifiedDeal                 bool            //required only for manual-bid deal
	FastRetrieval                bool            //required only for manual-bid deal
	SkipConfirmation             bool            //required only for manual-bid deal
	Duration                     int             //not necessary, when not provided use default value:1512000
	MinerPrice                   decimal.Decimal //used internally, not need to provide
	StartEpoch                   int             //required only for manual-bid deal
	StartEpochIntervalHours      int             //invalid
	OutputDir                    string          //required
	MinerFid                     string          //required only for manual-bid deal
	MetadataJsonPath             string          //required only for manual-bid deal
	DealSourceIds                []int           //required
	RelativeEpochFromMainNetwork int             //required
}
```

## Deal
```shell
type Deal struct {
	Uuid           string `json:"uuid"`
	SourceFileName string `json:"source_file_name"`
	MinerId        string `json:"miner_id"`
	DealCid        string `json:"deal_cid"`
	PayloadCid     string `json:"payload_cid"`
	FileSourceUrl  string `json:"file_source_url"`
	Md5            string `json:"md5"`
	StartEpoch     *int   `json:"start_epoch"`
	PieceCid       string `json:"piece_cid"`
	FileSize       int64  `json:"file_size"`
	Cost           string `json:"cost"`
}
```
