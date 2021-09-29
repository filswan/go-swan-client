package utils

import (
	"go-swan-client/logs"
	"go-swan-client/models"
	"regexp"
	"strings"
)

func LotusGetDealOnChainStatus(dealCid string) (string, string) {
	cmd := "lotus-miner storage-deals list -v | grep " + dealCid
	result, err := ExecOsCmd(cmd)

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
	result, err := ExecOsCmd(cmd)

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
	if words != nil && len(words) > 0 {
		currentEpoch = GetInt64FromStr(words[0])
	}

	logs.GetLogger().Info("currentEpoch: ", currentEpoch)
	return int(currentEpoch)
}

func LotusImportData(dealCid string, filepath string) string {
	cmd := "lotus-miner storage-deals import-data " + dealCid + " " + filepath
	logs.GetLogger().Info(cmd)

	result, err := ExecOsCmd(cmd)

	if err != nil {
		logs.GetLogger().Error(err)
		return ""
	}

	return result
}

func LotusGetMinerInfo(miner *models.Miner) bool {
	cmd := "lotus client query-ask " + miner.MinerFid
	logs.GetLogger().Info(cmd)

	result, err := ExecOsCmd(cmd)

	if err != nil {
		logs.GetLogger().Error(err)
		return false
	}

	if len(result) == 0 {
		logs.GetLogger().Error("Failed to get info for:", miner.MinerFid)
		return false
	}

	lines := strings.Split(result, "\n")
	logs.GetLogger().Info(lines)

	for _, line := range lines {
		if strings.Contains(line, "Verified Price per GiB:") {
			miner.VerifiedPrice = SearchFloat64FromStr(line)
			if miner.VerifiedPrice != nil {
				logs.GetLogger().Info("miner VerifiedPrice: ", *miner.VerifiedPrice)
			} else {
				logs.GetLogger().Error("Failed to get miner VerifiedPrice from lotus")
			}

			continue
		}

		if strings.Contains(line, "Price per GiB:") {
			miner.Price = SearchFloat64FromStr(line)
			if miner.Price != nil {
				logs.GetLogger().Info("miner Price: ", *miner.Price)
			} else {
				logs.GetLogger().Error("Failed to get miner Price from lotus")
			}

			continue
		}

		if strings.Contains(line, "Max Piece size:") {
			words := strings.Split(line, ":")
			if len(words) == 2 {
				maxPieceSize := strings.Trim(words[1], " ")
				miner.MaxPieceSize = &maxPieceSize
				if miner.MaxPieceSize != nil {
					logs.GetLogger().Info("miner MaxPieceSize: ", *miner.MaxPieceSize)
				} else {
					logs.GetLogger().Error("Failed to get miner MaxPieceSize from lotus")
				}
			}
			continue
		}

		if strings.Contains(line, "Min Piece size:") {
			words := strings.Split(line, ":")
			if len(words) == 2 {
				minPieceSize := strings.Trim(words[1], " ")
				miner.MinPieceSize = &minPieceSize
				if miner.MinPieceSize != nil {
					logs.GetLogger().Info("miner MinPieceSize: ", *miner.MinPieceSize)
				} else {
					logs.GetLogger().Error("Failed to get miner MinPieceSize from lotus")
				}
			}
			continue
		}
	}

	return true
}
