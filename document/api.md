# Groups

* [CreateCarFiles](#CreateCarFiles)
* [CreateGoCarFiles](#CreateGoCarFiles)
* [CreateIpfsCarFiles](#CreateIpfsCarFiles)
* [CreateIpfsCmdCarFiles](#CreateIpfsCmdCarFiles)
* [UploadCarFiles](#UploadCarFiles)
* [CreateTask](#CreateTask)
* [SendDeals](#SendDeals)
* [SendAutoBidDealsLoop](#SendAutoBidDealsLoop)
* [SendAutoBidDeals](#SendAutoBidDeals)
* [SendAutoBidDealsByTaskUuid](#SendAutoBidDealsByTaskUuid)

## CreateCarFiless

Definition:

```shell
func (cmdCar *CmdCar) CreateCarFiles() ([]*libmodel.FileDesc, error)
```

Outputs:

```shell
[]*libmodel.FileDesc  # files description
error                 # error or nil
```

## CreateGoCarFiles

Definition:

```shell
func (cmdGoCar *CmdGoCar) CreateGoCarFiles() ([]*libmodel.FileDesc, error)
```

Outputs:

```shell
[]*libmodel.FileDesc   # files description
error                  # error or nil
```

## CreateIpfsCarFiles

Definition:

```shell
func (cmdIpfsCar *CmdIpfsCar) CreateIpfsCarFiles() ([]*libmodel.FileDesc, error)
```

Outputs:

```shell
[]*libmodel.FileDesc   # files description
error                  # error or nil
```

## CreateIpfsCmdCarFiles

Definition:

```shell
func (cmdIpfsCmdCar *CmdIpfsCmdCar) CreateIpfsCmdCarFiles() ([]*libmodel.FileDesc, error)
```

Outputs:

```shell
[]*libmodel.FileDesc   # files description
error                  # error or nil
```

## UploadCarFiles

Definition:

```shell
(cmdUpload *CmdUpload) UploadCarFiles() ([]*libmodel.FileDesc, error)
```

Outputs:

```shell
[]*libmodel.FileDesc  # files description
error                 # error or nil
```

## CreateTask

Definition:

```shell
func (cmdTask *CmdTask) CreateTask(cmdDeal *CmdDeal) (*string, []*libmodel.FileDesc, []*Deal, error)
```

Inputs:

```shell
cmdDeal   # if you don't need to send deal, this can be nil
```

Outputs:

```shell
*string               # json file full path
[]*libmodel.FileDesc  # files description
error                 # error or nil
```

## SendDeals

Definition:

```shell
func (cmdDeal *CmdDeal) SendDeals() ([]*libmodel.FileDesc, error)
```

Outputs:

```shell
[]*libmodel.FileDesc  # files description
error                 # error or nil
```

## SendAutoBidDeals

Definition:

```shell
func (cmdAutoBidDeal *CmdAutoBidDeal) SendAutoBidDeals() ([][]*libmodel.FileDesc, error)
```

Outputs:

```shell
[][]*libmodel.FileDesc  # files description
error                   # error or nil
```

## SendAutoBidDealsByTaskUuid

Definition:

```shell
func (cmdAutoBidDeal *CmdAutoBidDeal) SendAutoBidDealsByTaskUuid(taskUuid string) (*string, []*libmodel.FileDesc, error)
```

Outputs:

```shell
*string               # json file path
[]*libmodel.FileDesc  # car files info
error                 # error or nil
```

## SendAutoBidDealsLoop

Definition:

```shell
func (cmdAutoBidDeal *CmdAutoBidDeal) SendAutoBidDealsLoop() error
```

Outputs:

```shell
error                   # error or nil
```
