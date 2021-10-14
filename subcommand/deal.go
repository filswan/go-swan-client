package subcommand

import (
	"go-swan-client/common/client"
	"go-swan-client/config"
	"go-swan-client/logs"
	"go-swan-client/model"
	"math"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

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

	result := SendDeals2Miner(nil, taskName, minerFid, *outputDir, carFiles)

	swanClient := client.SwanGetClient()
	response := swanClient.SwanUpdateTaskByUuid(carFiles[0].Uuid, minerFid, "")
	logs.GetLogger().Info(response)

	return result
}

func GetDealConfig(minerFid string) *model.DealConfig {
	dealConfig := model.DealConfig{
		MinerFid:           minerFid,
		SenderWallet:       config.GetConfig().Sender.Wallet,
		VerifiedDeal:       config.GetConfig().Sender.VerifiedDeal,
		FastRetrieval:      config.GetConfig().Sender.FastRetrieval,
		EpochIntervalHours: config.GetConfig().Sender.StartEpochHours,
		SkipConfirmation:   config.GetConfig().Sender.SkipConfirmation,
		StartEpochHours:    config.GetConfig().Sender.StartEpochHours,
	}

	maxPriceStr := config.GetConfig().Sender.MaxPrice
	maxPrice, err := decimal.NewFromString(maxPriceStr)
	if err != nil {
		logs.GetLogger().Error("Failed to convert maxPrice(" + maxPriceStr + ") to decimal, MaxPrice:")
		return nil
	}
	dealConfig.MaxPrice = maxPrice

	return &dealConfig
}

func CheckDealConfig(dealConfig *model.DealConfig) bool {
	minerPrice, minerVerifiedPrice, _, _ := client.LotusGetMinerConfig(dealConfig.MinerFid)

	if dealConfig.SenderWallet == "" {
		logs.GetLogger().Error("Sender.wallet should be set in config file.")
		return false
	}

	if dealConfig.VerifiedDeal {
		if minerVerifiedPrice == nil {
			return false
		}
		dealConfig.MinerPrice = *minerVerifiedPrice
		logs.GetLogger().Info("Miner price is:", *minerVerifiedPrice)
	} else {
		if minerPrice == nil {
			return false
		}
		dealConfig.MinerPrice = *minerPrice
		logs.GetLogger().Info("Miner price is:", *minerPrice)
	}

	logs.GetLogger().Info("Miner price is:", dealConfig.MinerPrice, " MaxPrice:", dealConfig.MaxPrice, " VerifiedDeal:", dealConfig.VerifiedDeal)
	priceCmp := dealConfig.MaxPrice.Cmp(dealConfig.MinerPrice)
	logs.GetLogger().Info("priceCmp:", priceCmp)
	if priceCmp < 0 {
		logs.GetLogger().Error("miner price is higher than deal max price")
		return false
	}

	logs.GetLogger().Info("Deal check passed.")

	return true
}

func SendDeals2Miner(dealConfig *model.DealConfig, taskName string, minerFid string, outputDir string, carFiles []*model.FileDesc) bool {
	if dealConfig == nil {
		dealConfig = GetDealConfig(minerFid)
		if dealConfig == nil {
			logs.GetLogger().Error("Failed to get deal config.")
			return false
		}
	}

	result := CheckDealConfig(dealConfig)
	if !result {
		logs.GetLogger().Error("Failed to pass deal config check.")
		return false
	}

	for _, carFile := range carFiles {
		if carFile.CarFileSize <= 0 {
			logs.GetLogger().Error("File:" + carFile.CarFilePath + " %s is too small")
			continue
		}
		pieceSize, sectorSize := CalculatePieceSize(carFile.CarFileSize)
		logs.GetLogger().Info("dealConfig.MinerPrice:", dealConfig.MinerPrice)
		cost := CalculateRealCost(sectorSize, dealConfig.MinerPrice)
		dealCid, startEpoch := client.LotusProposeOfflineDeal(cost, pieceSize, carFile.DataCid, carFile.PieceCid, *dealConfig)
		if dealCid == nil || startEpoch == nil {
			logs.GetLogger().Error("Failed to propose offline deal")
			return false
		}
		carFile.MinerFid = &minerFid
		carFile.DealCid = *dealCid
		carFile.StartEpoch = strconv.Itoa(*startEpoch)
	}

	jsonFileName := taskName + JSON_FILE_NAME_BY_DEAL_SUFFIX
	csvFileName := taskName + CSV_FILE_NAME_BY_DEAL_SUFFIX
	WriteCarFilesToFiles(carFiles, outputDir, jsonFileName, csvFileName)

	csvFilename := taskName + "_deal.csv"
	CreateCsv4TaskDeal(carFiles, &minerFid, outputDir, csvFilename)

	return true
}

// https://docs.filecoin.io/store/lotus/very-large-files/#maximizing-storage-per-sector
func CalculatePieceSize(fileSize int64) (int64, float64) {
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

func CalculateRealCost(sectorSizeBytes float64, pricePerGiB decimal.Decimal) decimal.Decimal {
	logs.GetLogger().Info("sectorSizeBytes:", sectorSizeBytes, " pricePerGiB:", pricePerGiB)
	bytesPerGiB := decimal.NewFromInt(1024 * 1024 * 1024)
	sectorSizeGiB := decimal.NewFromFloat(sectorSizeBytes).Div(bytesPerGiB)
	realCost := sectorSizeGiB.Mul(pricePerGiB)
	logs.GetLogger().Info("realCost:", realCost)
	return realCost
}
