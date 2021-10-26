# Groups
* [SubCommand](#SubCommand)
  * [GenerateCarFiles](#GenerateCarFiles)
  * [CreateGoCarFiles](#CreateGoCarFiles)
  * [CreateCarFilesDesc](#CreateCarFilesDesc)
  * [UploadCarFiles](#UploadCarFiles)
  * [CreateTask](#CreateTask)
  * [SendTask2Swan](#SendTask2Swan)
  * [SendDeals](#SendDeals)
  * [SendDeals2Miner](#SendDeals2Miner)
  * [SendAutoBidDeal](#SendAutoBidDeal)
  * [SendAutobidDeal](#SendAutobidDeal)
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


## SubCommand
### CreateCarFiless

Inputs:
```shell
confCar *model.ConfCar
```

Outputs:
```shell
[]*model.FileDesc
error
```

### CreateGoCarFiles

Inputs:
```shell
confCar *model.ConfCar
```

Outputs:
```shell
[]*model.FileDesc
error
```

### CreateCarFilesDescFromGoCarManifest

Inputs:
```shell
confCar *model.ConfCar
srcFileDir string
carFileDir string
```

Outputs:
```shell
[]*model.FileDesc, error
```

### UploadCarFiles

Inputs:
```shell
confUpload *model.ConfUpload
```

Outputs:
```shell
error
```

### CreateTask

Inputs:
```shell
confTask *model.ConfTask, confDeal *model.ConfDeal
```

Outputs:
```shell
*string, error
```

### SendTask2Swan

Inputs:
```shell
confTask *model.ConfTask, task model.Task, carFiles []*model.FileDesc
```

Outputs:
```shell
error
```

### SendDeals

Inputs:
```shell
confDeal *model.ConfDeal
```

Outputs:
```shell
error
```

### SendDeals2Miner

Inputs:
```shell
confDeal *model.ConfDeal, taskName string, outputDir string, carFiles []*model.FileDesc
```

Outputs:
```shell
*string, error
```

### SendAutoBidDeal

Inputs:
```shell
confDeal *model.ConfDeal
```

Outputs:
```shell
[]string, error
```

### SendAutobidDeal

Inputs:
```shell
confDeal *model.ConfDeal, deals []model.OfflineDeal, task model.Task, outputDir string
```

Outputs:
```shell
int, string, error
```

### CheckDealConfig

Inputs:
```shell
confDeal *model.ConfDeal
```

Outputs:
```shell
error
```

### CheckInputDir

Inputs:
```shell
inputDir string
```

Outputs:
```shell
error
```

### CreateOutputDir

Inputs:
```shell
outputDir string
```

Outputs:
```shell
error
```

### WriteCarFilesToFiles

Inputs:
```shell
confDeal *model.ConfDeal, deals []model.OfflineDeal, task model.Task, outputDir string
```

Outputs:
```shell
int, string, error
```

### WriteCarFilesToJsonFile

Inputs:
```shell
carFiles []*model.FileDesc, outputDir, jsonFilename, csvFileName string
```

Outputs:
```shell
error
```

### ReadCarFilesFromJsonFile

Inputs:
```shell
carFiles []*model.FileDesc, outputDir, jsonFilename string
```

Outputs:
```shell
error
```

### ReadCarFilesFromJsonFileByFullPath

Inputs:
```shell
confDeal *model.ConfDeal, deals []model.OfflineDeal, task model.Task, outputDir string
```

Outputs:
```shell
int, string, error
```

### CreateCsv4TaskDeal

Inputs:
```shell
confDeal *model.ConfDeal, deals []model.OfflineDeal, task model.Task, outputDir string
```

Outputs:
```shell
int, string, error
```

### WriteCarFilesToCsvFile

Inputs:
```shell
confDeal *model.ConfDeal, deals []model.OfflineDeal, task model.Task, outputDir string
```

Outputs:
```shell
int, string, error
```

### WriteCarFilesToCsvFile

Inputs:
```shell
confDeal *model.ConfDeal, deals []model.OfflineDeal, task model.Task, outputDir string
```

Outputs:
```shell
int, string, error
```

### WriteCarFilesToCsvFile

Inputs:
```shell
confDeal *model.ConfDeal, deals []model.OfflineDeal, task model.Task, outputDir string
```

Outputs:
```shell
int, string, error
```

