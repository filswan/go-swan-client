package subcommand

import (
	"errors"
	"math"
	"path/filepath"
	"strings"

	"github.com/filswan/go-swan-client/config"

	"github.com/filswan/go-swan-client/model"

	"github.com/filswan/go-swan-client/logs"

	"github.com/filswan/go-swan-client/common/constants"

	"github.com/filswan/go-swan-client/common/client"

	"github.com/shopspring/decimal"
)

func SendDeals(minerFid string, outputDir *string, metadataJsonPath string) bool {
	if outputDir == nil {
		outDir := config.GetConfig().Sender.OutputDir
		outputDir = &outDir
	}
	metadataJsonFilename := filepath.Base(metadataJsonPath)
	taskName := strings.TrimSuffix(metadataJsonFilename, constants.JSON_FILE_NAME_BY_TASK)
	carFiles := ReadCarFilesFromJsonFileByFullPath(metadataJsonPath)
	if carFiles == nil {
		logs.GetLogger().Error("Failed to read car files from json.")
		return false
	}

	csvFilepath, err := SendDeals2Miner(nil, taskName, minerFid, *outputDir, carFiles)
	if err != nil {
		logs.GetLogger().Error(err)
		return false
	}

	swanClient := client.SwanGetClient()
	response := swanClient.SwanUpdateTaskByUuid(carFiles[0].Uuid, minerFid, *csvFilepath)
	logs.GetLogger().Info(response)
	return true
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

func SendDeals2Miner(dealConfig *model.DealConfig, taskName string, minerFid string, outputDir string, carFiles []*model.FileDesc) (*string, error) {
	if dealConfig == nil {
		dealConfig = GetDealConfig(minerFid)
		if dealConfig == nil {
			err := errors.New("failed to get deal config")
			logs.GetLogger().Error(err)
			return nil, err
		}
	}

	err := CheckDealConfig(dealConfig)
	if err != nil {
		err := errors.New("failed to pass deal config check")
		logs.GetLogger().Error(err)
		return nil, err
	}

	for _, carFile := range carFiles {
		if carFile.CarFileSize <= 0 {
			logs.GetLogger().Error("File:" + carFile.CarFilePath + " %s is too small")
			continue
		}
		pieceSize, sectorSize := CalculatePieceSize(carFile.CarFileSize)
		logs.GetLogger().Info("dealConfig.MinerPrice:", dealConfig.MinerPrice)
		cost := CalculateRealCost(sectorSize, dealConfig.MinerPrice)
		dealCid, startEpoch, err := client.LotusProposeOfflineDeal(*carFile, cost, pieceSize, *dealConfig, 0)
		//dealCid, err := client.LotusClientStartDeal(*carFile, cost, pieceSize, *dealConfig)
		if err != nil {
			logs.GetLogger().Error(err)
			continue
		}
		if dealCid == nil {
			continue
		}
		carFile.MinerFid = &dealConfig.MinerFid
		carFile.DealCid = *dealCid
		carFile.StartEpoch = *startEpoch

		logs.GetLogger().Info("Cid:", carFile.DealCid)
	}

	jsonFileName := taskName + constants.JSON_FILE_NAME_BY_DEAL
	csvFileName := taskName + constants.CSV_FILE_NAME_BY_DEAL
	err = WriteCarFilesToFiles(carFiles, outputDir, jsonFileName, csvFileName)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	csvFilename := taskName + "-deals.csv"
	csvFilepath, err := CreateCsv4TaskDeal(carFiles, outputDir, csvFilename)

	return &csvFilepath, err
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
