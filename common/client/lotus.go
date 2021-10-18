package client

import (
	"bufio"
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
	dataCid = strings.TrimRight(dataCid, "\n")

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

func LotusProposeOfflineDeal(carFile model.FileDesc, cost decimal.Decimal, pieceSize int64, dealConfig model.DealConfig) (*string, error) {
	fastRetrieval := strings.ToLower(strconv.FormatBool(dealConfig.FastRetrieval))
	verifiedDeal := strings.ToLower(strconv.FormatBool(dealConfig.VerifiedDeal))
	costFloat, _ := cost.Float64()
	costStr := fmt.Sprintf("%.18f", costFloat)
	logs.GetLogger().Info("wallet:", dealConfig.SenderWallet)
	logs.GetLogger().Info("miner:", dealConfig.MinerFid)
	logs.GetLogger().Info("price:", dealConfig.MinerPrice)
	logs.GetLogger().Info("total cost:", costStr)
	logs.GetLogger().Info("start epoch:", carFile.StartEpoch)
	logs.GetLogger().Info("fast-retrieval:", fastRetrieval)
	logs.GetLogger().Info("verified-deal:", verifiedDeal)

	cmd := "lotus client deal --from " + dealConfig.SenderWallet
	cmd = cmd + " --start-epoch " + strconv.Itoa(carFile.StartEpoch)
	cmd = cmd + " --fast-retrieval=" + fastRetrieval + " --verified-deal=" + verifiedDeal
	cmd = cmd + " --manual-piece-cid " + carFile.PieceCid + " --manual-piece-size " + strconv.FormatInt(pieceSize, 10)
	cmd = cmd + " " + carFile.DataCid + " " + dealConfig.MinerFid + " " + costStr + " " + strconv.Itoa(constants.DURATION)
	logs.GetLogger().Info(cmd)

	if !dealConfig.SkipConfirmation {
		logs.GetLogger().Info("Do you confirm to submit the deal? Press Y/y to continue, other key to quit")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}

		response = strings.TrimRight(response, "\n")

		if strings.ToUpper(response) != "Y" {
			logs.GetLogger().Info("Your input is ", response, ". Now give up submit the deal.")
			return nil, nil
		}
	}

	result, err := ExecOsCmd(cmd, true)

	if err != nil {
		logs.GetLogger().Error("Failed to submit the deal.")
		logs.GetLogger().Error(err)
		return nil, err
	}
	result = strings.Trim(result, "\n")
	logs.GetLogger().Info(result)

	dealCid := result

	return &dealCid, nil
}
