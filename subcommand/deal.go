package subcommand

import (
	"fmt"
	"go-swan-client/common/client"
	"go-swan-client/config"
	"go-swan-client/logs"
	"go-swan-client/model"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const DURATION = "1051200"
const EPOCH_PER_HOUR = 120

type DealConfig struct {
	MinerFid           string
	SenderWallet       string
	MaxPrice           float64
	VerifiedDeal       bool
	FastRetrieval      bool
	EpochIntervalHours int
	SkipConfirmation   bool
	Price              float64
}

func SendDeals(minerFid string, outputDir *string, metadataJsonPath string) bool {
	if outputDir == nil {
		outDir := config.GetConfig().Sender.OutputDir
		outputDir = &outDir
	}
	metadataJsonFilename := filepath.Base(metadataJsonPath)
	taskName := strings.TrimSuffix(metadataJsonFilename, JSON_FILE_NAME_BY_TASK_SUFFIX)
	carFiles := ReadCarFilesFromJsonFileByFullPath(metadataJsonPath)
	if carFiles == nil {
		logs.GetLogger().Error("Failed to read car files from json.")
		return false
	}

	result := SendDeals2Miner(taskName, minerFid, *outputDir, carFiles)

	return result
}

func GetDealConfig(minerFid string) *DealConfig {
	dealConfig := DealConfig{
		MinerFid:           minerFid,
		SenderWallet:       config.GetConfig().Sender.Wallet,
		VerifiedDeal:       config.GetConfig().Sender.VerifiedDeal,
		FastRetrieval:      config.GetConfig().Sender.FastRetrieval,
		EpochIntervalHours: config.GetConfig().Sender.StartEpochHours,
		SkipConfirmation:   config.GetConfig().Sender.SkipConfirmation,
	}
	maxPriceStr := config.GetConfig().Sender.MaxPrice
	maxPrice, err := strconv.ParseFloat(maxPriceStr, 64)
	if err == nil {
		logs.GetLogger().Error("Failed to convert maxPrice to float.")
		return nil
	}
	dealConfig.MaxPrice = maxPrice

	return &dealConfig
}

func CheckDealConfig(dealConfig DealConfig) bool {
	minerPrice, minerVerifiedPrice, _, _ := client.LotusGetMinerConfig(dealConfig.MinerFid)

	if config.GetConfig().Sender.VerifiedDeal {
		if minerVerifiedPrice == nil {
			return false
		}
		dealConfig.Price = *minerVerifiedPrice
	} else {
		if minerPrice == nil {
			return false
		}
		dealConfig.Price = *minerPrice
	}

	if dealConfig.Price > dealConfig.MaxPrice {
		msg := fmt.Sprintf("miner %s price %f higher than max price %f", dealConfig.MinerFid, dealConfig.Price, dealConfig.MaxPrice)
		logs.GetLogger().Error(msg)
		return false
	}

	return true
}

func SendDeals2Miner(taskName string, minerFid string, outputDir string, carFiles []*model.FileDesc) bool {
	dealConfig := GetDealConfig(minerFid)
	if dealConfig == nil {
		logs.GetLogger().Error("Failed to get deal config.")
		return false
	}

	result := CheckDealConfig(*dealConfig)
	if !result {
		logs.GetLogger().Error("Failed to pass deal config check.")
		return false
	}

	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		logs.GetLogger().Error(err)
		return false
	}

	for _, carFile := range carFiles {
		if carFile.CarFileSize <= 0 {
			msg := fmt.Sprintf("File %s is too small", carFile.CarFilePath)
			logs.GetLogger().Error(msg)
			continue
		}
		pieceSize, sectorSize := calculatePieceSizeFromFileSize(carFile.CarFileSize)
		cost := calculateRealCost(sectorSize, dealConfig.Price)
		dealCid, startEpoch := client.LotusProposeOfflineDeal(dealConfig.Price, cost, pieceSize, carFile.DataCid, carFile.PieceCid, minerFid)
		outputCsvPath := ""
		carFile.MinerFid = &minerFid
		carFile.DealCid = *dealCid
		carFile.StartEpoch = strconv.Itoa(*startEpoch)

		logs.GetLogger().Info("Swan deal final CSV Generated: %s", outputCsvPath)

	}

	jsonFileName := taskName + JSON_FILE_NAME_BY_DEAL_SUFFIX
	csvFileName := taskName + CSV_FILE_NAME_BY_DEAL_SUFFIX
	WriteCarFilesToFiles(carFiles, outputDir, jsonFileName, csvFileName)

	return true
}

// https://docs.filecoin.io/store/lotus/very-large-files/#maximizing-storage-per-sector
func calculatePieceSizeFromFileSize(fileSize int64) (int64, float64) {
	exp := math.Ceil(math.Log2(float64(fileSize)))
	sectorSize2Check := math.Pow(2, exp)
	pieceSize2Check := int64(sectorSize2Check * 254 / 256)
	if fileSize <= pieceSize2Check {
		return pieceSize2Check, sectorSize2Check
	}

	exp = exp + 1
	realSectorSize := math.Pow(2, exp)
	realPieceSize := int64(realSectorSize * 254 / 256)
	return realPieceSize, realSectorSize
}

func calculateRealCost(sectorSizeBytes float64, pricePerGiB float64) float64 {
	var bytesPerGiB float64 = 1024 * 1024 * 1024
	sectorSizeGiB := float64(sectorSizeBytes) / bytesPerGiB

	realCost := sectorSizeGiB * pricePerGiB
	return realCost
}
