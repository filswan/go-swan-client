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
	Result ClientCalcCommPResult `json:"result"`
}

type ClientCalcCommPResult struct {
	Root DealCid
	Size int
}
type DealCid struct {
	DealCid string `json:"/"`
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

//"lotus-miner storage-deals list -v | grep -a " + dealCid
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

func LotusClientCalcCommP(filepath string) *MarketGetAskResultAsk {
	lotusClient := LotusGetClient()

	var params []interface{}

	jsonRpcParams := LotusJsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_MARKET_GET_ASK,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	response := HttpPost(lotusClient.MinerApiUrl, lotusClient.AccessToken, jsonRpcParams)

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
