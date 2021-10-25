package utils

import (
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go-swan-client/logs"

	"github.com/shopspring/decimal"
)

// GetEpochInMillis get current timestamp
func GetEpochInMillis() (millis int64) {
	nanos := time.Now().UnixNano()
	millis = nanos / 1000000
	return
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

func UrlJoin(root string, parts ...string) string {
	url := root

	for _, part := range parts {
		url = strings.TrimRight(url, "/") + "/" + strings.TrimLeft(part, "/")
	}
	url = strings.TrimRight(url, "/")

	return url
}
