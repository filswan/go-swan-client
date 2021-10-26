# Groups
* [Ipfs](#Ipfs)
  * [IpfsUploadCarFile](#IpfsUploadCarFile)
* [Lotus](#Lotus)
  * [LotusGetClient](#LotusGetClient)
  * [LotusClientCalcCommP](#LotusClientCalcCommP)
  * [LotusClientImport](#LotusClientImport)
  * [LotusClientGenCar](#LotusClientGenCar)
  * [LotusGetMinerConfig()](#LotusGetMinerConfigs())
  * [LotusProposeOfflineDeal](#LotusProposeOfflineDeal)
* [ExecOsCmd](#ExecOsCmd)
  * [ExecOsCmd2Screen](#ExecOsCmd2Screen)
  * [ExecOsCmd](#ExecOsCmd)
  * [ExecOsCmdBase](#ExecOsCmdBase)
* [Http](#[Http)
  * [HttpPostNoToken](#HttpPostNoToken)
  * [HttpPost](#HttpPost)
  * [HttpGetNoToken](#HttpGetNoToken)
  * [HttpGet](#HttpGets)
  * [HttpPut](#HttpPut)
  * [HttpDelete](#HttpDelete)
  * [httpRequest](#httpRequest)
  * [HttpPutFile](#HttpPutFile)
  * [HttpPostFile](#HttpPostFile)
  * [HttpRequestFile](#HttpRequestFile)
* [Swan](#Swan)
  * [SwanGetJwtToken](#SwanGetJwtToken)
  * [SwanGetClient](#SwanGetClient)
  * [SwanGetOfflineDeals](#SwanGetOfflineDeal)
  * [SwanUpdateOfflineDealStatus](#SwanUpdateOfflineDealStatus)
  * [SwanCreateTask](#SwanCreateTask)
  * [SwanGetTasks](#SwanGetTasks)
  * [SwanGetAssignedTasks](#SwanGetAssignedTasks)
  * [SwanGetOfflineDealsByTaskUuid](#SwanGetOfflineDealsByTaskUuid)
  * [SwanUpdateTaskByUuid](#SwanUpdateTaskByUuid)
  * [SwanUpdateAssignedTask](#SwanUpdateAssignedTask)


## Ipfs
### IpfsUploadCarFile

Inputs:
```shell
carFilePath string
```

Outputs:
```shell
*string: car file hash
error: error or nil
```

## Lotus
### LotusGetClients

Inputs:
```shell
apiUrl  string   #lotus node api url, such as http://[ip]:[port]/rpc/v0
accessToken  string  #lotus node access token, should have admin privilege
```

Outputs:
```shell
*LotusClient #structure including access info for lotus node
error: error or nil
```

### LotusClientCalcCommP

Inputs:
```shell
filepath string
```

Outputs:
```shell
*string  #piece cid, or nil when cannot get the info required
```

### LotusClientImport

Inputs:
```shell
filepath string
isCar bool
```

Outputs:
```shell
*string  #piece cid, or nil when cannot get the info required
```

### LotusClientGenCar

Inputs:
```shell
srcFilePath string
destCarFilePath string
srcFilePathIsCar bool
```

Outputs:
```shell
error  #error or nils
```

### LotusGetMinerConfig

Inputs:
```shell
minerFid string
```

Outputs:
```shell
*decimal.Decimal  # price
*decimal.Decimal  # verified price
*string  # max piece size
*string  # min piece size
```

### LotusProposeOfflineDeal

Inputs:
```shell
carFile model.FileDesc
cost decimal.Decimal
pieceSize int64
dealConfig model.ConfDeal
relativeEpoch int
```

Outputs:
```shell
*string  # deal cid
*int  # start epoch
error # error or nil
```

## ExecOsCmd
### ExecOsCmd2Screen

Inputs:
```shell
cmdStr string
checkStdErr bool
```

Outputs:
```shell
string  # standard output
error # error or nil
```

### ExecOsCmd

Inputs:
```shell
cmdStr string
checkStdErr bool
```

Outputs:
```shell
string  # standard output
error # error or nil
```


### ExecOsCmdBase

Inputs:
```shell
cmdStr string
out2Screen bool
checkStdErr bool
```

Outputs:
```shell
string  # standard output
error # error or nil
```

## Http
### HttpPostNoToken

Inputs:
```shell
uri string
params interface{}
```

Outputs:
```shell
string  # result from web api request, if error, then ""
```

### HttpPost

Inputs:
```shell
uri
tokenString string
params interface{}
```

Outputs:
```shell
string  # result from web api request, if error, then ""
```

### HttpGetNoToken

Inputs:
```shell
uri string
params interface{}
```

Outputs:
```shell
string  # result from web api request, if error, then ""
```

### HttpGet

Inputs:
```shell
uri
tokenString string
params interface{}
```

Outputs:
```shell
string  # result from web api request, if error, then ""
```

### HttpPut

Inputs:
```shell
uri
tokenString string
params interface{}
```

Outputs:
```shell
string  # result from web api request, if error, then ""
```

### HttpDelete

Inputs:
```shell
uri
tokenString string
params interface{}
```

Outputs:
```shell
string  # result from web api request, if error, then ""
```

### httpRequest

Inputs:
```shell
httpMethod string
uri string
tokenString string
params interface{}
```

Outputs:
```shell
string  # result from web api request, if error, then ""
```

### HttpPutFile

Inputs:
```shell
url string
tokenString string
paramTexts map[string]string
paramFilename
paramFilepath string
```

Outputs:
```shell
string  # result from web api request, if error, then ""
error # error or nil
```

### HttpPostFile

Inputs:
```shell
url string
tokenString string
paramTexts map[string]string
paramFilename
paramFilepath string
```

Outputs:
```shell
string  # result from web api request, if error, then ""
error # error or nil
```

### HttpRequestFile

Inputs:
```shell
httpMethod string
url string string
tokenString string
paramTexts map[string]string
paramFilename string
paramFilepath string
```

Outputs:
```shell
string  # result from web api request, if error, then ""
error # error or nil
```

## Swan
### SwanGetJwtToken

Inputs:
```shell
httpMethod string
url string string
tokenString string
paramTexts map[string]string
paramFilename string
paramFilepath string
```

Outputs:
```shell
string  # result from web api request, if error, then ""
error # error or nil
```


### SwanGetJwtToken

Inputs:
```shell
apiKey string
accessToken string
```

Outputs:
```shell
error # error or nil
```

### SwanGetClient

Inputs:
```shell
apiUrl string
apiKey string
accessToken string
```

Outputs:
```shell
*SwanClient
error
```

### SwanGetOfflineDeals

Inputs:
```shell
minerFid string
status string
limit ...string
```

Outputs:
```shell
[]model.OfflineDeal
```

### SwanUpdateOfflineDealStatus

Inputs:
```shell
dealId int
status string
statusInfo ...string
```

Outputs:
```shell
bool
```

### SwanCreateTask

Inputs:
```shell
task model.Task
csvFilePath string
```

Outputs:
```shell
*SwanCreateTaskResponse
error
```

### SwanGetTasks

Inputs:
```shell
limit *int
```

Outputs:
```shell
*GetTaskResult
error
```

### SwanGetAssignedTasks

Inputs:
```shell
```

Outputs:
```shell
[]model.Task
error
```

### SwanGetOfflineDealsByTaskUuid

Inputs:
```shell
taskUuid string
```

Outputs:
```shell
*GetOfflineDealsByTaskUuidResult
error
```

### SwanUpdateTaskByUuid

Inputs:
```shell
taskUuid string
minerFid string
csvFilePath string
```

Outputs:
```shell
*GetOfflineDealsByTaskUuidResult
error
```

### SwanUpdateAssignedTask

Inputs:
```shell
taskUuid
status
csvFilePath string
```

Outputs:
```shell
*SwanCreateTaskResponse
error
```
