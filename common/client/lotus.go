package client

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"go-swan-client/common/constants"
	"go-swan-client/common/utils"
	"go-swan-client/logs"
	"go-swan-client/model"

	"github.com/shopspring/decimal"
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
	ApiUrl      string
	AccessToken string
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

func LotusGetClient(apiUrl, accessToken string) (*LotusClient, error) {
	if len(apiUrl) == 0 {
		err := fmt.Errorf("config lotus api_url is required")
		logs.GetLogger().Error(err)
		return nil, err
	}

	if len(accessToken) == 0 {
		err := fmt.Errorf("config lotus access_token is required and it should have admin access right")
		logs.GetLogger().Error(err)
		return nil, err
	}

	lotusClient := &LotusClient{
		ApiUrl:      apiUrl,
		AccessToken: accessToken,
	}

	return lotusClient, nil
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
func (lotusClient *LotusClient) LotusVersion() (*string, error) {
	var params []interface{}

	jsonRpcParams := LotusJsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_VERSION,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	//here the api url should be miner's api url, need to change later on
	response := HttpGetNoToken(lotusClient.ApiUrl, jsonRpcParams)
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
func (lotusClient *LotusClient) LotusMarketGetAsk() *MarketGetAskResultAsk {
	var params []interface{}

	jsonRpcParams := LotusJsonRpcParams{
		JsonRpc: LOTUS_JSON_RPC_VERSION,
		Method:  LOTUS_MARKET_GET_ASK,
		Params:  params,
		Id:      LOTUS_JSON_RPC_ID,
	}

	//here the api url should be miner's api url, need to change later on
	response := HttpGetNoToken(lotusClient.ApiUrl, jsonRpcParams)
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
func (lotusClient *LotusClient) LotusClientCalcCommP(filepath string) *string {
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
func (lotusClient *LotusClient) LotusClientImport(filepath string, isCar bool) (*string, error) {
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
		err := fmt.Errorf("lotus import file %s failed, no response from %s", filepath, lotusClient.ApiUrl)
		logs.GetLogger().Error(err)
		return nil, err
	}

	clientImport := &ClientImport{}
	err := json.Unmarshal([]byte(response), clientImport)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if clientImport.Error != nil {
		err := fmt.Errorf("lotus import file %s failed, error code:%d, message:%s", filepath, clientImport.Error.Code, clientImport.Error.Message)
		logs.GetLogger().Error(err)
		return nil, err
	}

	if clientImport.Result == nil {
		err := fmt.Errorf("lotus import file %s failed, result is null from %s", filepath, lotusClient.ApiUrl)
		logs.GetLogger().Error(err)
		return nil, err
	}

	dataCid := clientImport.Result.Root.Cid

	return &dataCid, nil
}

//"lotus client generate-car " + srcFilePath + " " + destCarFilePath
func (lotusClient *LotusClient) LotusClientGenCar(srcFilePath, destCarFilePath string, srcFilePathIsCar bool) error {
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
		err := fmt.Errorf("failed to generate car, no response")
		logs.GetLogger().Error(err)
		return err
	}

	logs.GetLogger().Info(response)
	lotusJsonRpcResult := &LotusJsonRpcResult{}
	err := json.Unmarshal([]byte(response), lotusJsonRpcResult)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	if lotusJsonRpcResult.Error != nil {
		err := fmt.Errorf("error, code:%d, message:%s", lotusJsonRpcResult.Error.Code, lotusJsonRpcResult.Error.Message)
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
func (lotusClient *LotusClient) LotusClientStartDeal(carFile model.FileDesc, cost decimal.Decimal, pieceSize int64, dealConfig model.ConfDeal) (*string, error) {
	//costFloat, _ := cost.Float64()
	//costStr := fmt.Sprintf("%.18f", costFloat)

	var params []interface{}
	clientStartDealParamData := ClientStartDealParamData{
		TransferType: "string value",
		Root: Cid{
			Cid: carFile.DataCid,
		},
		PieceCid: Cid{
			Cid: carFile.PieceCid,
		},
		PieceSize:    int(pieceSize),
		RawBlockSize: 42,
	}
	clientStartDealParam := ClientStartDealParam{
		Data:               clientStartDealParamData,
		Wallet:             dealConfig.SenderWallet,
		Miner:              *dealConfig.MinerFid,
		EpochPrice:         "2",
		MinBlocksDuration:  constants.DURATION,
		ProviderCollateral: "0",
		DealStartEpoch:     carFile.StartEpoch,
		FastRetrieval:      dealConfig.FastRetrieval,
		VerifiedDeal:       dealConfig.VerifiedDeal,
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
		err := fmt.Errorf("failed to generate car, no response")
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
		err := fmt.Errorf("error, code:%d, message:%s", clientStartDeal.Error.Code, clientStartDeal.Error.Message)
		logs.GetLogger().Error(err)
		return nil, err
	}

	logs.GetLogger().Info("Cid:", clientStartDeal.Result.Cid)
	return &clientStartDeal.Result.Cid, nil
}

func LotusGetMinerConfig(minerFid string) (*decimal.Decimal, *decimal.Decimal, *string, *string) {
	cmd := "lotus client query-ask " + minerFid
	logs.GetLogger().Info(cmd)

	result, err := ExecOsCmd(cmd, true)

	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, nil, nil
	}

	if len(result) == 0 {
		logs.GetLogger().Error("Failed to get info for:", minerFid)
		return nil, nil, nil, nil
	}

	lines := strings.Split(result, "\n")
	logs.GetLogger().Info(lines)

	var verifiedPrice *decimal.Decimal
	var price *decimal.Decimal
	var maxPieceSize string
	var minPieceSize string
	for _, line := range lines {
		if strings.Contains(line, "Verified Price per GiB:") {
			verifiedPrice, err = utils.GetDecimalFromStr(line)
			if err != nil {
				logs.GetLogger().Error("Failed to get miner VerifiedPrice from lotus")
			} else {
				logs.GetLogger().Info("miner verifiedPrice: ", *verifiedPrice)
			}

			continue
		}

		if strings.Contains(line, "Price per GiB:") {
			price, err = utils.GetDecimalFromStr(line)
			if err != nil {
				logs.GetLogger().Error("Failed to get miner Price from lotus")
			} else {
				logs.GetLogger().Info("miner Price: ", *price)
			}

			continue
		}

		if strings.Contains(line, "Max Piece size:") {
			words := strings.Split(line, ":")
			if len(words) == 2 {
				maxPieceSize = strings.Trim(words[1], " ")
				if maxPieceSize != "" {
					logs.GetLogger().Info("miner MaxPieceSize: ", maxPieceSize)
				} else {
					logs.GetLogger().Error("Failed to get miner MaxPieceSize from lotus")
				}
			}
			continue
		}

		if strings.Contains(line, "Min Piece size:") {
			words := strings.Split(line, ":")
			if len(words) == 2 {
				minPieceSize = strings.Trim(words[1], " ")
				if minPieceSize != "" {
					logs.GetLogger().Info("miner MinPieceSize: ", minPieceSize)
				} else {
					logs.GetLogger().Error("Failed to get miner MinPieceSize from lotus")
				}
			}
			continue
		}
	}

	return price, verifiedPrice, &maxPieceSize, &minPieceSize
}

func LotusProposeOfflineDeal(carFile model.FileDesc, cost decimal.Decimal, pieceSize int64, dealConfig model.ConfDeal, relativeEpoch int) (*string, *int, error) {
	fastRetrieval := strings.ToLower(strconv.FormatBool(dealConfig.FastRetrieval))
	verifiedDeal := strings.ToLower(strconv.FormatBool(dealConfig.VerifiedDeal))
	costFloat, _ := cost.Float64()
	costStr := fmt.Sprintf("%.18f", costFloat)
	startEpoch := dealConfig.StartEpoch - relativeEpoch

	logs.GetLogger().Info("wallet:", dealConfig.SenderWallet)
	logs.GetLogger().Info("miner:", dealConfig.MinerFid)
	logs.GetLogger().Info("price:", dealConfig.MinerPrice)
	logs.GetLogger().Info("total cost:", costStr)
	logs.GetLogger().Info("start epoch:", startEpoch)
	logs.GetLogger().Info("fast-retrieval:", fastRetrieval)
	logs.GetLogger().Info("verified-deal:", verifiedDeal)

	cmd := "lotus client deal --from " + dealConfig.SenderWallet
	cmd = cmd + " --start-epoch " + strconv.Itoa(startEpoch)
	cmd = cmd + " --fast-retrieval=" + fastRetrieval + " --verified-deal=" + verifiedDeal
	cmd = cmd + " --manual-piece-cid " + carFile.PieceCid + " --manual-piece-size " + strconv.FormatInt(pieceSize, 10)
	cmd = cmd + " " + carFile.DataCid + " " + *dealConfig.MinerFid + " " + costStr + " " + strconv.Itoa(constants.DURATION)
	logs.GetLogger().Info(cmd)

	if !dealConfig.SkipConfirmation {
		logs.GetLogger().Info("Do you confirm to submit the deal?")
		logs.GetLogger().Info("Press Y/y to continue, other key to quit")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, nil, err
		}

		response = strings.TrimRight(response, "\n")

		if strings.ToUpper(response) != "Y" {
			logs.GetLogger().Info("Your input is ", response, ". Now give up submit the deal.")
			return nil, nil, nil
		}
	}

	result, err := ExecOsCmd(cmd, true)

	if err != nil {
		logs.GetLogger().Error("Failed to submit the deal.")
		logs.GetLogger().Error(err)
		return nil, nil, err
	}
	result = strings.Trim(result, "\n")
	logs.GetLogger().Info(result)

	dealCid := result

	return &dealCid, &startEpoch, nil
}
