package utils

import (
	"errors"
	"fmt"
	"go-swan-client/common/constants"
	"go-swan-client/config"
	"go-swan-client/logs"
	"regexp"
	"strconv"
	"strings"
)

func LotusGetDealOnChainStatus(dealCid string) (string, string) {
	cmd := "lotus-miner storage-deals list -v | grep " + dealCid
	result, err := ExecOsCmd(cmd, true)

	if err != nil {
		logs.GetLogger().Error("Failed to get deal on chain status, please check if lotus-miner is running properly.")
		logs.GetLogger().Error(err)
		return "", ""
	}

	if len(result) == 0 {
		logs.GetLogger().Error("Deal does not found on chain. DealCid:", dealCid)
		return "", ""
	}

	words := strings.Fields(result)
	status := ""
	for _, word := range words {
		if strings.HasPrefix(word, "StorageDeal") {
			status = word
			break
		}
	}

	if len(status) == 0 {
		return "", ""
	}

	message := ""

	for i := 11; i < len(words); i++ {
		message = message + words[i] + " "
	}

	message = strings.TrimRight(message, " ")
	return status, message
}

func LotusGetCurrentEpoch() int {
	cmd := "lotus-miner proving info | grep 'Current Epoch'"
	logs.GetLogger().Info(cmd)
	result, err := ExecOsCmd(cmd, true)

	if err != nil {
		logs.GetLogger().Error(err)
		return -1
	}

	if len(result) == 0 {
		logs.GetLogger().Error("Failed to get current epoch. Please check if miner is running properly.")
		return -1
	}

	logs.GetLogger().Info(result)

	re := regexp.MustCompile("[0-9]+")
	words := re.FindAllString(result, -1)
	logs.GetLogger().Info("words:", words)
	var currentEpoch int64 = -1
	if len(words) > 0 {
		currentEpoch = GetInt64FromStr(words[0])
	}

	logs.GetLogger().Info("currentEpoch: ", currentEpoch)
	return int(currentEpoch)
}

func LotusImportData(dealCid string, filepath string) string {
	cmd := "lotus-miner storage-deals import-data " + dealCid + " " + filepath
	logs.GetLogger().Info(cmd)

	result, err := ExecOsCmd(cmd, true)

	if err != nil {
		logs.GetLogger().Error(err)
		return ""
	}

	return result
}

func LotusGetMinerConfig(minerFid string) (*float64, *float64, *string, *string) {
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

	var verifiedPrice *float64
	var price *float64
	var maxPieceSize string
	var minPieceSize string
	for _, line := range lines {
		if strings.Contains(line, "Verified Price per GiB:") {
			verifiedPrice = SearchFloat64FromStr(line)
			if verifiedPrice != nil {
				logs.GetLogger().Info("miner verifiedPrice: ", *verifiedPrice)
			} else {
				logs.GetLogger().Error("Failed to get miner VerifiedPrice from lotus")
			}

			continue
		}

		if strings.Contains(line, "Price per GiB:") {
			price = SearchFloat64FromStr(line)
			if price != nil {
				logs.GetLogger().Info("miner Price: ", *price)
			} else {
				logs.GetLogger().Error("Failed to get miner Price from lotus")
			}

			continue
		}

		if strings.Contains(line, "Max Piece size:") {
			words := strings.Split(line, ":")
			if len(words) == 2 {
				maxPieceSize = strings.Trim(words[1], " ")
				if maxPieceSize == "" {
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
				if minPieceSize == "" {
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

func LotusGeneratePieceCid(carFilePath string) *string {
	cmd := "lotus client commP " + carFilePath
	logs.GetLogger().Info(cmd)

	result, err := ExecOsCmd(cmd, true)

	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	if len(result) == 0 {
		logs.GetLogger().Error("Failed to get info for:", carFilePath)
		return nil
	}

	lines := strings.Split(result, "\n")
	logs.GetLogger().Info(lines)

	var pieceCid *string
	for _, line := range lines {
		if strings.Contains(line, "CID:") {
			words := strings.Fields(line)
			if len(words) < 2 {
				return nil
			}
			fileCid := strings.Trim(words[1], " ")
			pieceCid = &fileCid
			continue
		}
	}

	if pieceCid == nil {
		logs.GetLogger().Error("Cannot get file cid:", carFilePath)
		return nil
	}

	logs.GetLogger().Info("pieceCid:", *pieceCid)

	return pieceCid
}

func LotusImportCarFile(carFilePath string) *string {
	cmd := "lotus client import --car " + carFilePath
	logs.GetLogger().Info(cmd)

	result, err := ExecOsCmd(cmd, true)

	if err != nil {
		logs.GetLogger().Error(err)
		return nil
	}

	if len(result) == 0 {
		logs.GetLogger().Error("Failed to import:", carFilePath)
		return nil
	}

	words := strings.Split(result, " Root ")
	if len(words) < 2 {
		logs.GetLogger().Error("Failed to import:", carFilePath)
		return nil
	}

	dataCid := strings.Trim(words[1], " ")

	return &dataCid
}

func LotusGenerateCar(srcFilePath, destCarFilePath string) error {
	cmd := "lotus client generate-car " + srcFilePath + " " + destCarFilePath
	logs.GetLogger().Info(cmd)

	result, err := ExecOsCmd(cmd, true)

	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	if len(result) != 0 {
		errMsg := fmt.Sprintf("Generate car file %s for %s failed", destCarFilePath, srcFilePath)
		err = errors.New(errMsg)
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}

func LotusProposeOfflineDeal(price, cost float64, pieceSize int64, dataCid, pieceCid, minerId string) (*string, *int) {
	epochIntervalHours := config.GetConfig().Sender.StartEpochHours
	fromWallet := config.GetConfig().Sender.Wallet
	startEpoch := GetCurrentEpoch() + (epochIntervalHours+1)*constants.EPOCH_PER_HOUR
	fastRetrieval := strings.ToLower(strconv.FormatBool(config.GetConfig().Sender.FastRetrieval))
	verifiedDeal := strings.ToLower(strconv.FormatBool(config.GetConfig().Sender.VerifiedDeal))
	cmd := "lotus client deal --from " + fromWallet + " --start-epoch " + strconv.Itoa(startEpoch) + " --fast-retrieval=" + fastRetrieval + " --verified-deal=" + verifiedDeal
	cmd = cmd + " --manual-piece-cid " + pieceCid + " --manual-piece-size " + strconv.FormatInt(pieceSize, 10) + " " + dataCid + " " + minerId + " " + strconv.FormatFloat(cost, 'f', -1, 10)

	logs.GetLogger().Info(cmd)
	logs.GetLogger().Info("wallet:", fromWallet)
	logs.GetLogger().Info("miner:", minerId)
	logs.GetLogger().Info("price:", price)
	logs.GetLogger().Info("total cost:", cost)
	logs.GetLogger().Info("start epoch:", startEpoch)
	logs.GetLogger().Info("fast-retrieval:", fastRetrieval)
	logs.GetLogger().Info("verified-deal:", verifiedDeal)

	skipConfirmation := config.GetConfig().Sender.SkipConfirmation
	if !skipConfirmation {
		logs.GetLogger().Info("Press Enter to continue...")
		var response string
		_, err := fmt.Scanln(&response)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, nil
		}

		if response != "/n" {
			return nil, nil
		}
	}

	result, err := ExecOsCmd(cmd, false)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil
	}

	result = strings.Trim(result, " ")
	dealCid := result

	return &dealCid, &startEpoch
}
