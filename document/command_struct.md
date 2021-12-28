# Groups

* [CmdCar](#CmdCar)
* [CmdGoCar](#CmdGoCar)
* [CmdIpfsCar](#CmdIpfsCar)
* [CmdUpload](#CmdUpload)
* [CmdTask](#CmdTask)
* [CmdDeal](#CmdDeal)
* [CmdAutoBidDeal](#CmdAutoBidDeal)
* [Deal](#Deal)

## CmdCar
```shell
type CmdCar struct {
	LotusClientApiUrl      string //required
	LotusClientAccessToken string //required
	OutputDir              string //required
	InputDir               string //required
	GenerateMd5            bool   //required
}
```

## CmdGoCar
```shell
type CmdGoCar struct {
	LotusClientApiUrl      string //required
	LotusClientAccessToken string //required
	OutputDir              string //required
	InputDir               string //required
	GenerateMd5            bool   //required
	GocarFileSizeLimit     int64  //required
	GocarFolderBased       bool   //required
}
```

## CmdIpfsCar
```shell
type CmdIpfsCar struct {
	LotusClientApiUrl         string //required
	LotusClientAccessToken    string //required
	OutputDir                 string //required
	InputDir                  string //required
	GenerateMd5               bool   //required
	IpfsServerUploadUrlPrefix string //required
}

```

## CmdUpload
```shell
type CmdUpload struct {
	StorageServerType           string //required
	IpfsServerDownloadUrlPrefix string //required only when upload to ipfs server
	IpfsServerUploadUrlPrefix   string //required only when upload to ipfs server
	OutputDir                   string //invalid
	InputDir                    string //required
}
```

## CmdTask
```shell
type CmdTask struct {
	SwanApiUrl                 string          //required when OfflineMode is false
	SwanApiKey                 string          //required when OfflineMode is false and SwanJwtToken is not provided
	SwanAccessToken            string          //required when OfflineMode is false and SwanJwtToken is not provided
	SwanToken                  string          //required when OfflineMode is false and SwanApiKey & SwanAccessToken are not provided
	LotusClientApiUrl          string          //required
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
	Dataset                    string          //not necessary
	Description                string          //not necessary
	StartEpochHours            int             //required
	SourceId                   int             //required
	MaxAutoBidCopyNumber       int             //required only for public autobid deal
}
```

## CmdDeal
```shell
type CmdDeal struct {
	SwanApiUrl             string          //required
	SwanApiKey             string          //required when SwanJwtToken is not provided
	SwanAccessToken        string          //required when SwanJwtToken is not provided
	SwanToken              string          //required when SwanApiKey and SwanAccessToken are not provided
	LotusClientApiUrl      string          //required
	LotusClientAccessToken string          //required
	SenderWallet           string          //required
	MaxPrice               decimal.Decimal //required
	VerifiedDeal           bool            //required
	FastRetrieval          bool            //required
	SkipConfirmation       bool            //required
	Duration               int             //not necessary, when not provided use default value:1512000
	StartEpochHours        int             //required
	OutputDir              string          //required
	MinerFids              []string        //required
	MetadataJsonPath       string          //required
	DealSourceIds          []int           //required
}
```

## CmdAutoBidDeal
```shell
type CmdAutoBidDeal struct {
	SwanApiUrl             string //required
	SwanApiKey             string //required when SwanJwtToken is not provided
	SwanAccessToken        string //required when SwanJwtToken is not provided
	SwanToken              string //required when SwanApiKey and SwanAccessToken are not provided
	LotusClientApiUrl      string //required
	LotusClientAccessToken string //required
	SenderWallet           string //required
	OutputDir              string //required
	DealSourceIds          []int  //required
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
