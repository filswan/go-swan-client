package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-swan-client/common/constants"
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
	LOTUS_CLIENT_GEN_CAR         = "Filecoin.ClientGenCar"
	LOTUS_CLIENT_START_DEAL      = "Filecoin.ClientStartDeal"
	LOTUS_VERSION                = "Filecoin.Version"
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

type LotusVersionResult struct {
	Version    string
	APIVersion int
	BlockDelay int
}

type LotusVersionResponse struct {
	LotusJsonRpcResult
	Result LotusVersionResult `json:"result"`
}

//"lotus client query-ask " + minerFid
func LotusVersion() (*string, error) {
	lotusClient := LotusGetClient()

	var params []interface{}

	jsonRpcParams := LotusJsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_VERSION,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	response := HttpGetNoToken(lotusClient.MinerApiUrl, jsonRpcParams)
	if response == "" {
		err := errors.New("no response from api")
		logs.GetLogger().Error(err)
		return nil, err
	}

	lotusVersionResponse := &LotusVersionResponse{}
	err := json.Unmarshal([]byte(response), lotusVersionResponse)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if lotusVersionResponse.Error != nil {
		msg := fmt.Sprintf("error, code:%d, message:%s", lotusVersionResponse.Error.Code, lotusVersionResponse.Error.Message)
		err := errors.New(msg)
		logs.GetLogger().Error(err)
		return nil, err
	}

	return &lotusVersionResponse.Result.Version, nil
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

type ClientFileParam struct {
	Path  string
	IsCAR bool
}

//"lotus client import --car " + carFilePath
func LotusClientImport(filepath string, isCar bool) *string {
	lotusClient := LotusGetClient()

	var params []interface{}
	clientFileParam := ClientFileParam{
		Path:  filepath,
		IsCAR: isCar,
	}
	params = append(params, clientFileParam)

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

//"lotus client generate-car " + srcFilePath + " " + destCarFilePath
func LotusClientGenCar(srcFilePath, destCarFilePath string, srcFilePathIsCar bool) error {
	lotusClient := LotusGetClient()

	var params []interface{}
	clientFileParam := ClientFileParam{
		Path:  srcFilePath,
		IsCAR: srcFilePathIsCar,
	}
	params = append(params, clientFileParam)
	params = append(params, destCarFilePath)

	jsonRpcParams := LotusJsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_CLIENT_GEN_CAR,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	response := HttpGet(lotusClient.ApiUrl, lotusClient.AccessToken, jsonRpcParams)
	if response == "" {
		err := errors.New("failed to generate car, no response")
		logs.GetLogger().Error(err)
		return err
	}

	lotusJsonRpcResult := &LotusJsonRpcResult{}
	err := json.Unmarshal([]byte(response), lotusJsonRpcResult)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	if lotusJsonRpcResult.Error != nil {
		msg := fmt.Sprintf("error, code:%d, message:%s", lotusJsonRpcResult.Error.Code, lotusJsonRpcResult.Error.Message)
		err := errors.New(msg)
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}

type ClientStartDealParamData struct {
	TransferType string
	Root         Cid
	PieceCid     Cid
	PieceSize    int
	RawBlockSize int
}

type ClientStartDealParam struct {
	Data               ClientStartDealParamData
	Wallet             string
	Miner              string
	EpochPrice         string
	MinBlocksDuration  int
	ProviderCollateral string
	DealStartEpoch     int
	FastRetrieval      bool
	VerifiedDeal       bool
}

type ClientStartDeal struct {
	LotusJsonRpcResult
	Result *Cid `json:"result"`
}

//"lotus client generate-car " + srcFilePath + " " + destCarFilePath
func LotusClientStartDeal(minerFid, dataCid, pieceCid, wallet, epochPrice string, pieceSize, startEpoch int, fastRetrieval, verifiedDeal bool) (*string, error) {
	lotusClient := LotusGetClient()

	var params []interface{}
	clientStartDealParamData := ClientStartDealParamData{
		TransferType: "string value",
		Root: Cid{
			Cid: dataCid,
		},
		PieceCid: Cid{
			Cid: pieceCid,
		},
		PieceSize:    pieceSize,
		RawBlockSize: 42,
	}
	clientStartDealParam := ClientStartDealParam{
		Data:               clientStartDealParamData,
		Wallet:             wallet,
		Miner:              minerFid,
		EpochPrice:         epochPrice,
		MinBlocksDuration:  constants.DURATION,
		ProviderCollateral: "0",
		DealStartEpoch:     startEpoch,
		FastRetrieval:      fastRetrieval,
		VerifiedDeal:       verifiedDeal,
	}
	params = append(params, clientStartDealParam)

	jsonRpcParams := LotusJsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_CLIENT_START_DEAL,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	response := HttpGet(lotusClient.ApiUrl, lotusClient.AccessToken, jsonRpcParams)
	if response == "" {
		err := errors.New("failed to generate car, no response")
		logs.GetLogger().Error(err)
		return nil, err
	}

	clientStartDeal := &ClientStartDeal{}
	err := json.Unmarshal([]byte(response), clientStartDeal)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if clientStartDeal.Error != nil {
		msg := fmt.Sprintf("error, code:%d, message:%s", clientStartDeal.Error.Code, clientStartDeal.Error.Message)
		err := errors.New(msg)
		logs.GetLogger().Error(err)
		return nil, err
	}

	return &clientStartDeal.Result.Cid, nil
}
