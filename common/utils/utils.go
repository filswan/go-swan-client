package utils

import (
	"context"
	"encoding/json"
	"go-swan-client/logs"
	"io/ioutil"
	"math/big"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// GetEpochInMillis get current timestamp
func GetEpochInMillis() (millis int64) {
	nanos := time.Now().UnixNano()
	millis = nanos / 1000000
	return
}

func ReadContractAbiJsonFile(aptpath string) (string, error) {
	jsonFile, err := os.Open(aptpath)

	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}

	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}
	return string(byteValue), nil
}

func GetRewardPerBlock() *big.Int {
	rewardBig, _ := new(big.Int).SetString("35000000000000000000", 10) // the unit is wei
	return rewardBig
}

func CheckTx(client *ethclient.Client, tx *types.Transaction) (*types.Receipt, error) {
retry:
	rp, err := client.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		if err == ethereum.NotFound {
			logs.GetLogger().Error("tx ", tx.Hash().String(), " not found, check it later")
			time.Sleep(1 * time.Second)
			goto retry
		} else {
			logs.GetLogger().Error("TransactionReceipt fail: ", err)
			return nil, err
		}
	}
	return rp, nil
}

func GetFromAndToAddressByTxHash(client *ethclient.Client, chainID *big.Int, txHash common.Hash) (*addressInfo, error) {
	addrInfo := new(addressInfo)
	tx, _, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	addrInfo.AddrTo = tx.To().Hex()
	txMsg, err := tx.AsMessage(types.NewEIP155Signer(chainID), nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}
	addrInfo.AddrFrom = txMsg.From().Hex()
	return addrInfo, nil
}

type addressInfo struct {
	AddrFrom string
	AddrTo   string
}

func ToJson(obj interface{}) (string, error) {
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}

	jsonString := string(jsonBytes)
	return jsonString, nil
}

func GetInt64FromStr(numStr string) int64 {
	num, err := strconv.ParseInt(numStr, 10, 64)
	if err != nil {
		logs.GetLogger().Error(err)
		return -1
	}

	return num
}

func GetFloat64FromStr(numStr *string) (float64, error) {
	if numStr == nil || *numStr == "" {
		return -1, nil
	}

	*numStr = strings.Trim(*numStr, " ")
	if *numStr == "" {
		return -1, nil
	}

	num, err := strconv.ParseFloat(*numStr, 64)
	if err != nil {
		logs.GetLogger().Error(err)
		return -1, err
	}

	return num, nil
}

func GetIntFromStr(numStr string) (int, error) {
	num, err := strconv.ParseInt(numStr, 10, 32)
	if err != nil {
		logs.GetLogger().Error(err)
		return -1, err
	}

	return int(num), nil
}

func GetNumStrFromStr(numStr string) string {
	re := regexp.MustCompile("[0-9]+.?[0-9]*")
	words := re.FindAllString(numStr, -1)
	//logs.GetLogger().Info("words:", words)
	if words != nil && len(words) > 0 {
		return words[0]
	}

	return ""
}

func GetByteSizeFromStr(sizeStr string) *float64 {
	sizeStr = strings.Trim(sizeStr, " ")
	numStr := GetNumStrFromStr(sizeStr)
	numStr = strings.Trim(numStr, " ")
	size, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}
	unit := strings.Trim(sizeStr, numStr)
	unit = strings.Trim(unit, " ")
	if len(unit) == 0 {
		return &size
	}

	unit = strings.ToUpper(unit)
	switch unit {
	case "GIB", "GB":
		size = size * 1024 * 1024 * 1024
	case "MIB", "MB":
		size = size * 1024 * 1024
	case "KIB", "KB":
		size = size * 1024
	case "BYTE", "B":
		return &size
	default:
		return nil
	}

	return &size
}

func IsSameDay(nanoSec1, nanoSec2 int64) bool {
	year1, month1, day1 := time.Unix(0, nanoSec1).Date()
	year2, month2, day2 := time.Unix(0, nanoSec2).Date()

	if year1 == year2 && month1 == month2 && day1 == day2 {
		return true
	}

	return false
}

func GetRandInRange(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	randVal := min + rand.Intn(max-min+1)
	return randVal
}

func IsStrEmpty(str *string) bool {
	if str == nil || *str == "" {
		return true
	}

	strTrim := strings.Trim(*str, " ")
	if len(strTrim) == 0 {
		return true
	}

	return false
}

func GetDayNumFromEpoch(epoch int) int {
	return epoch / 2 / 60 / 24
}

func GetEpochFromDay(day int) int {
	return day * 24 * 60 * 2
}

func GetMinFloat64(val1, val2 *float64) *float64 {
	if val1 == nil {
		return val2
	}

	if val2 == nil {
		return val1
	}

	if *val1 <= *val2 {
		return val1
	}

	return val2
}

func GetCurrentEpoch() int {
	currentNanoSec := time.Now().UnixNano()
	currentEpoch := (currentNanoSec/1e9 - 1598306471) / 30
	return int(currentEpoch)
}

func GetFieldStrFromJson(jsonStr string, fieldName string) string {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		logs.GetLogger().Error(err)
		return ""
	}

	fieldVal := result[fieldName].(interface{})
	return fieldVal.(string)
}

func GetFieldMapFromJson(jsonStr string, fieldName string) map[string]interface{} {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	fieldVal := result[fieldName].(interface{})

	return fieldVal.(map[string]interface{})
}

func SearchFloat64FromStr(source string) *float64 {
	re := regexp.MustCompile("[0-9]*.?[0-9]*")
	words := re.FindAllString(source, -1)
	logs.GetLogger().Info("words:", words)
	if words != nil && len(words) > 0 {
		result, err := strconv.ParseFloat(words[0], 64)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil
		}
		return &result
	}

	return nil
}
