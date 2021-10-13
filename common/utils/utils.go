package utils

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"go-swan-client/logs"
	"io"
	"io/ioutil"
	"math/big"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
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
	if len(words) > 0 {
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
	return len(strTrim) == 0
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

	fieldVal := result[fieldName]
	return fieldVal.(string)
}

func GetFieldMapFromJson(jsonStr string, fieldName string) map[string]interface{} {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	fieldVal := result[fieldName]

	return fieldVal.(map[string]interface{})
}

func IsFileExists(dir, fileName string) bool {
	fileFullPath := filepath.Join(dir, fileName)
	_, err := os.Stat(fileFullPath)

	if err != nil {
		logs.GetLogger().Info(err)
		return false
	}

	return true
}

func IsFileExistsFullPath(fileFullPath string) bool {
	_, err := os.Stat(fileFullPath)

	if err != nil {
		logs.GetLogger().Info(err)
		return false
	}

	return true
}

func RemoveFile(dir, fileName string) {
	fileFullPath := filepath.Join(dir, fileName)
	err := os.Remove(fileFullPath)
	if err != nil {
		logs.GetLogger().Error(err.Error())
	}
}

func GetFileSize(fileFullPath string) int64 {
	fi, err := os.Stat(fileFullPath)
	if err != nil {
		logs.GetLogger().Info(err)
		return -1
	}

	return fi.Size()
}

func GetFileSize2(dir, fileName string) int64 {
	fileFullPath := filepath.Join(dir, fileName)
	fi, err := os.Stat(fileFullPath)
	if err != nil {
		logs.GetLogger().Info(err)
		return -1
	}

	return fi.Size()
}

func ReadAllLines(dir, filename string) ([]string, error) {
	fileFullPath := filepath.Join(dir, filename)

	file, err := os.Open(fileFullPath)

	if err != nil {
		logs.GetLogger().Error("failed opening file: ", fileFullPath)
		return nil, err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	lines := []string{}

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, nil
}

func ReadFile(filePath string) (string, []byte, error) {
	sourceFileStat, err := os.Stat(filePath)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", nil, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		err = errors.New(filePath + " is not a regular file")
		logs.GetLogger().Error(err)
		return "", nil, err
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		logs.GetLogger().Error("failed reading data from file: ", filePath)
		logs.GetLogger().Error(err)
		return "", nil, err
	}

	return sourceFileStat.Name(), data, nil
}

func copy(srcFilePath, destDir string) (int64, error) {
	sourceFileStat, err := os.Stat(srcFilePath)
	if err != nil {
		logs.GetLogger().Error(err)
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		err = errors.New(srcFilePath + " is not a regular file")
		logs.GetLogger().Error(err)
		return 0, err
	}

	source, err := os.Open(srcFilePath)
	if err != nil {
		logs.GetLogger().Error(err)
		return 0, err
	}

	defer source.Close()

	destination, err := os.Create(destDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return 0, err
	}

	defer destination.Close()

	nBytes, err := io.Copy(destination, source)
	if err != nil {
		logs.GetLogger().Error(err)
	}

	return nBytes, err
}

func GetDecimalFromStr(source string) (*decimal.Decimal, error) {
	re := regexp.MustCompile("[0-9]+.?[0-9]*")
	words := re.FindAllString(source, -1)
	if len(words) > 0 {
		numStr := strings.Trim(words[0], " ")
		result, err := decimal.NewFromString(numStr)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}
		return &result, nil
	}

	return nil, nil
}
