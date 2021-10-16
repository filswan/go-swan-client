package client

import (
	"encoding/json"
	"go-swan-client/config"
	"go-swan-client/logs"
)

const (
	LOTUS_JSON_RPC_ID            = 7878
	LOTUS_JSON_RPC_VERSION       = "2.0"
	LOTUS_CLIENT_GET_DEAL_INFO   = "Filecoin.ClientGetDealInfo"
	LOTUS_CLIENT_GET_DEAL_STATUS = "Filecoin.ClientGetDealStatus"
	LOTUS_CHAIN_HEAD             = "Filecoin.ChainHead"
	LOTUS_MARKET_GET_ASK         = "Filecoin.MarketGetAsk"
	LOTUS_CLIENT_CALC_COMM_P     = "Filecoin.ClientCalcCommP"
	LOTUS_CLIENT_IMPORT          = "Filecoin.ClientImport"
)

type LotusJsonRpcParams struct {
	JsonRpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      int           `json:"id"`
}

type LotusClient struct {
	ApiUrl           string
	AccessToken      string
	MinerApiUrl      string
	MinerAccessToken string
}

type LotusJsonRpcResult struct {
	Id      int           `json:"id"`
	JsonRpc string        `json:"jsonrpc"`
	Error   *JsonRpcError `json:"error"`
}

type MarketGetAsk struct {
	LotusJsonRpcResult
	Result *MarketGetAskResult `json:"result"`
}

type JsonRpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type MarketGetAskResult struct {
	Ask MarketGetAskResultAsk
}
type MarketGetAskResultAsk struct {
	Price         string
	VerifiedPrice string
	MinPieceSize  int
	MaxPieceSize  int
	Miner         string
	Timestamp     int
	Expiry        int
	SeqNo         int
}

type ClientCalcCommP struct {
	LotusJsonRpcResult
	Result *ClientCalcCommPResult `json:"result"`
}

type ClientCalcCommPResult struct {
	Root Cid
	Size int
}
type Cid struct {
	Cid string `json:"/"`
}

type ClientImport struct {
	LotusJsonRpcResult
	Result *ClientImportResult `json:"result"`
}
type ClientImportResult struct {
	Root     Cid
	ImportID int64
}

func LotusGetClient() *LotusClient {
	lotusClient := &LotusClient{
		ApiUrl:           config.GetConfig().Lotus.ApiUrl,
		AccessToken:      config.GetConfig().Lotus.AccessToken,
		MinerApiUrl:      config.GetConfig().Lotus.MinerApiUrl,
		MinerAccessToken: config.GetConfig().Lotus.MinerAccessToken,
	}

	return lotusClient
}

//"lotus client query-ask " + minerFid
func LotusMarketGetAsk() *MarketGetAskResultAsk {
	lotusClient := LotusGetClient()

	var params []interface{}

	jsonRpcParams := LotusJsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_MARKET_GET_ASK,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	response := HttpGetNoToken(lotusClient.MinerApiUrl, jsonRpcParams)
	if response == "" {
		return nil
	}

	marketGetAsk := &MarketGetAsk{}
	err := json.Unmarshal([]byte(response), marketGetAsk)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	if marketGetAsk.Result == nil {
		return nil
	}

	return &marketGetAsk.Result.Ask
}

//"lotus client commP " + carFilePath
func LotusClientCalcCommP(filepath string) *string {
	lotusClient := LotusGetClient()

	var params []interface{}
	params = append(params, filepath)

	jsonRpcParams := LotusJsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_CLIENT_CALC_COMM_P,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	response := HttpPost(lotusClient.ApiUrl, lotusClient.AccessToken, jsonRpcParams)
	if response == "" {
		return nil
	}

	clientCalcCommP := &ClientCalcCommP{}
	err := json.Unmarshal([]byte(response), clientCalcCommP)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	if clientCalcCommP.Result == nil {
		return nil
	}

	pieceCid := clientCalcCommP.Result.Root.Cid
	return &pieceCid
}

type ClientImportParam struct {
	Path  string
	IsCAR bool
}

//"lotus client import --car " + carFilePath
func LotusClientImport(filepath string, isCar bool) *string {
	lotusClient := LotusGetClient()

	var params []interface{}
	clientImportParam := ClientImportParam{
		Path:  filepath,
		IsCAR: isCar,
	}
	params = append(params, clientImportParam)

	jsonRpcParams := LotusJsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_CLIENT_IMPORT,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	response := HttpGet(lotusClient.ApiUrl, lotusClient.AccessToken, jsonRpcParams)
	if response == "" {
		return nil
	}

	clientImport := &ClientImport{}
	err := json.Unmarshal([]byte(response), clientImport)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	if clientImport.Result == nil {
		return nil
	}

	dataCid := clientImport.Result.Root.Cid

	return &dataCid
}
