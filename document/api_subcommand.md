# Groups
* [Car](#Car)
  * [GenerateCarFiles](#GenerateCarFiles)
* [GoCar](#GoCar)
  * [CreateGoCarFiles](#CreateGoCarFiles)
  * [CreateCarFilesDesc](#CreateCarFilesDesc)
* [Upload](#Upload)
  * [UploadCarFiles](#UploadCarFiles)
* [Task](#Task)
  * [CreateTask](#CreateTask)
  * [SendTask2Swan](#SendTask2Swan)
* [Deal](#Car)
  * [SendDeals](#SendDeals)
  * [SendDeals2Miner](#SendDeals2Miner)
* [Auto-Bid Deal](#Auto-Bid-Deal)
  * [SendAutoBidDeal](#SendAutoBidDeal)
  * [SendAutobidDeal](#SendAutobidDeal)
* [Common](#Common)
  * [CheckDealConfig](#CheckDealConfig)
  * [CheckInputDir](#CheckInputDir)
  * [CreateOutputDir](#CreateOutputDir)
  * [WriteCarFilesToFiles](#WriteCarFilesToFiles)
  * [WriteCarFilesToJsonFile](#WriteCarFilesToJsonFile)
  * [ReadCarFilesFromJsonFile](#ReadCarFilesFromJsonFile)
  * [ReadCarFilesFromJsonFileByFullPath](#ReadCarFilesFromJsonFileByFullPath)
  * [CreateCsv4TaskDeal](#CreateCsv4TaskDeal)
  * [WriteCarFilesToCsvFile](#WriteCarFilesToCsvFile)
  * [WriteCarFilesToCsvFile](#WriteCarFilesToCsvFile)
  * [WriteCarFilesToCsvFile](#WriteCarFilesToCsvFile)
  * [CalculatePieceSize](#CalculatePieceSize)
  * [CalculateRealCost](#CalculateRealCost)


## Car
### CreateCarFiless

Definition:
```shell
func CreateCarFiles(confCar *model.ConfCar) ([]*model.FileDesc, error)
```

Outputs:
```shell
[]*model.FileDesc   # car files info
error      # error or nil
```

## GoCar
### CreateGoCarFiles

Definition:
```shell
func CreateGoCarFiles(confCar *model.ConfCar) ([]*model.FileDesc, error) 
```

Outputs:
```shell
[]*model.FileDesc   # car files info
error      # error or nil
```

### CreateCarFilesDescFromGoCarManifest

Definition:
```shell
func CreateCarFilesDescFromGoCarManifest(confCar *model.ConfCar, srcFileDir, carFileDir string) ([]*model.FileDesc, error)
```

Outputs:
```shell
[]*model.FileDesc   # car files info
error      # error or nil
```

## Upload
### UploadCarFiles

Definition:
```shell
func UploadCarFiles(confUpload *model.ConfUpload) error
```

Outputs:
```shell
error   # error or nil
```

## Task
### CreateTask

Definition:
```shell
func CreateTask(confTask *model.ConfTask, confDeal *model.ConfDeal) (*string, error)
```

Outputs:
```shell
*string   # json filename
error     # error or nil
```

### SendTask2Swan

Definition:
```shell
func SendTask2Swan(confTask *model.ConfTask, task model.Task, carFiles []*model.FileDesc) error 
```

Outputs:
```shell
error  # error or nil
```

## Deal
### SendDeals

Definition:
```shell
func SendDeals(confDeal *model.ConfDeal) error 
```

Outputs:
```shell
error   # error or nil
```

### SendDeals2Miner

Definition:
```shell
func SendDeals2Miner(confDeal *model.ConfDeal, taskName string, outputDir string, carFiles []*model.FileDesc) (*string, error)
```

Outputs:
```shell
*string   # csvFilepath
error     # error or nil
```

## Auto-bid Deal
### SendAutoBidDeal

Definition:
```shell
func SendAutoBidDeal(confDeal *model.ConfDeal) ([]string, error)
```

Outputs:
```shell
[]string  #csvFilepaths
error     # error or nil
```

### SendAutobidDeal

Definition:
```shell
func SendAutobidDeal(confDeal *model.ConfDeal, deals []model.OfflineDeal, task model.Task, outputDir string) (int, string, error) 
```

Outputs:
```shell
int    # dealSentNum
string # csvFilepath
error  # error or nil
```

## Common
### CheckDealConfig

Definition:
```shell
func CheckDealConfig(confDeal *model.ConfDeal) error 
```

Outputs:
```shell
error   # error or nil
```

### CheckInputDir

Definition:
```shell
func CheckInputDir(inputDir string) error
```

Outputs:
```shell
error   # error or nil
```

### CreateOutputDir

Definition:
```shell
func CreateOutputDir(outputDir string) error 
```

Outputs:
```shell
error   # error or nil
```

### WriteCarFilesToFiles

Definition:
```shell
func WriteCarFilesToFiles(carFiles []*model.FileDesc, outputDir, jsonFilename, csvFileName string) error 
```

Outputs:
```shell
error   # error or nil
```

### WriteCarFilesToJsonFile

Definition:
```shell
func WriteCarFilesToJsonFile(carFiles []*model.FileDesc, outputDir, jsonFilename string) error 
```

Outputs:
```shell
error   # error or nil
```

### ReadCarFilesFromJsonFile

Definition:
```shell
func ReadCarFilesFromJsonFile(inputDir, jsonFilename string) []*model.FileDesc
```

Outputs:
```shell
[]*model.FileDesc  # car file info
```

### ReadCarFilesFromJsonFileByFullPath

Definition:
```shell
func ReadCarFilesFromJsonFileByFullPath(jsonFilePath string) []*model.FileDesc
```

Outputs:
```shell
[]*model.FileDesc  # car file info
```


### WriteCarFilesToCsvFile

Definition:
```shell
func WriteCarFilesToCsvFile(carFiles []*model.FileDesc, outDir, csvFileName string) error
```

Outputs:
```shell
error  # error or nil
```

### CreateCsv4TaskDeal

Definition:
```shell
func CreateCsv4TaskDeal(carFiles []*model.FileDesc, outDir, csvFileName string) (string, error)
```

Outputs:
```shell
string  # csv filepath
error   # error or nil
```
### CalculatePieceSize

Definition:
```shell
func CalculatePieceSize(fileSize int64) (int64, float64) 
```

Outputs:
```shell
int64   # piece size
float64 # sector size
```

### CalculateRealCost

Definition:
```shell
func CalculateRealCost(sectorSizeBytes float64, pricePerGiB decimal.Decimal) decimal.Decimal 
```

Outputs:
```shell
decimal.Decimal # real cost
```

