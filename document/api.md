# Groups
* [Client](#Client)
  * [IpfsUploadCarFile](#IpfsUploadCarFile)
  * [LotusGetClient](#LotusGetClient)
  * [LotusClientCalcCommP](#LotusClientCalcCommP)
  * [LotusClientImport](#LotusClientImport)
  * [LotusClientGenCar](#LotusClientGenCar)
  * [LotusGetMinerConfig()](#LotusGetMinerConfigs())
  * [LotusProposeOfflineDeal](#LotusProposeOfflineDeal)
  * [ExecOsCmd2Screen](#ExecOsCmd2Screen)
  * [ExecOsCmd](#ExecOsCmd)
  * [ExecOsCmdBase](#ExecOsCmdBase)
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
* [Beacon](#Beacon)
  * [BeaconGetEntry](#BeaconGetEntry)
## 


## Client
### IpfsUploadCarFile

Inputs:
```json
carFilePath string
```

Outputs:
```shell
*string: car file hash
error: error or nil
```

### LotusGetClients

Inputs:
```json
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
```json
filepath string
```

Outputs:
```shell
*string  #piece cid, or nil when cannot get the info required
```


### LotusClientImport

Inputs:
```json
filepath string
isCar bool
```

Outputs:
```shell
*string  #piece cid, or nil when cannot get the info required
```


### LotusClientGenCar

Inputs:
```json
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
```json
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
```json
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


### ExecOsCmd2Screen

Inputs:
```json
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
```json
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
```json
cmdStr string
out2Screen bool
checkStdErr bool
```

Outputs:
```shell
string  # standard output
error # error or nil
```

### HttpPostNoToken

Inputs:
```json
uri string
params interface{}
```

Outputs:
```shell
string  # result from web api request, if error, then ""
```

### HttpPost

Inputs:
```json
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
```json
uri string
params interface{}
```

Outputs:
```shell
string  # result from web api request, if error, then ""
```

### HttpGet

Inputs:
```json
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
```json
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
```json
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
```json
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
```json
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
```json
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
```json
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
```json
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

### SwanGetClient
### SwanGetOfflineDeals
### SwanUpdateOfflineDealStatus
### SwanCreateTask
### SwanGetTasks
### SwanGetAssignedTasks
### SwanGetOfflineDealsByTaskUuid
### SwanUpdateTaskByUuid
### SwanUpdateAssignedTask

