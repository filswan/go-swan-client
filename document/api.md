# Groups
* [GenerateCarFiles](#GenerateCarFiles)
* [CreateGoCarFiles](#CreateGoCarFiles)
* [UploadCarFiles](#UploadCarFiles)
* [CreateTask](#CreateTask)
* [SendDeals](#SendDeals)
* [SendAutoBidDeals](#SendAutoBidDeals)

## CreateCarFiless

Definition:
```shell
func CreateCarFiles(confCar *model.ConfCar) ([]*libmodel.FileDesc, error)
```

Outputs:
```shell
[]*libmodel.FileDesc  # car files info
error                 # error or nil
```

## CreateGoCarFiles

Definition:
```shell
func CreateGoCarFiles(confCar *model.ConfCar) ([]*libmodel.FileDesc, error)
```

Outputs:
```shell
[]*libmodel.FileDesc   # car files info
error                  # error or nil
```

## UploadCarFiles

Definition:
```shell
func UploadCarFiles(confUpload *model.ConfUpload) ([]*libmodel.FileDesc, error)
```

Outputs:
```shell
[]*libmodel.FileDesc  # car files info
error                 # error or nil
```

## CreateTask

Definition:
```shell
func CreateTask(confTask *model.ConfTask, confDeal *model.ConfDeal) (*string, []*libmodel.FileDesc, error)
```

Inputs:
```shell
confTask
confDeal   # if you don't need to send deal, this can be nil
```

Outputs:
```shell
*string               # json file full path
[]*libmodel.FileDesc  # car files info
error                 # error or nil
```

## SendDeals

Definition:
```shell
func SendDeals(confDeal *model.ConfDeal) ([]*libmodel.FileDesc, error)
```

Outputs:
```shell
[]*libmodel.FileDesc  # car files info
error                 # error or nil
```

## SendAutoBidDeal

Definition:
```shell
func SendAutoBidDeals(confDeal *model.ConfDeal) ([]string, [][]*libmodel.FileDesc, error)
```

Outputs:
```shell
[]string  #csvFilepaths
[]*libmodel.FileDesc  # car files info
error                 # error or nil
```

