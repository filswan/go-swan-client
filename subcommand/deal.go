package subcommand

import (
	"encoding/csv"
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
	SenderWallet       string `json:"sender_wallet"`
	MaxPrice           string `json:"max_price"`
	VerifiedDeal       bool   `json:"verified_deal"`
	FastRetrieval      bool   `json:"fast_retrieval"`
	EpochIntervalHours int    `json:"epoch_interval_hours"`
	SkipConfirmation   bool   `json:"skip_confirmation"`
}

func SendDeals(minerFid string, outputDir *string, metadataJsonPath string) bool {
	if outputDir == nil {
		outDir := filepath.Dir(metadataJsonPath)
		outputDir = &outDir
	}
	filename := filepath.Base(metadataJsonPath)
	taskName := strings.TrimSuffix(filename, filepath.Ext(filename))
	carFiles := ReadCarFilesFromJsonFileByFullPath(metadataJsonPath)
	if carFiles == nil {
		logs.GetLogger().Error("Failed to read car files from json.")
		return false
	}

	result := SendDeals2Miner(taskName, minerFid, *outputDir, carFiles)

	return result
}

func SendDeals2Miner(taskName string, minerFid string, outputDir string, carFiles []*model.FileDesc) bool {
	//dealConfig := DealConfig{
	//	SenderWallet:       config.GetConfig().Sender.Wallet,
	//	MaxPrice:           config.GetConfig().Sender.MaxPrice,
	//	VerifiedDeal:       config.GetConfig().Sender.VerifiedDeal,
	//	FastRetrieval:      config.GetConfig().Sender.FastRetrieval,
	//	EpochIntervalHours: config.GetConfig().Sender.StartEpochHours,
	//	SkipConfirmation:   config.GetConfig().Sender.SkipConfirmation,
	//}    if csv_file_path:

	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		logs.GetLogger().Error(err)
		return false
	}

	for _, carFile := range carFiles {
		//dataCid := carFile.DataCid
		//pieceCid := carFile.PieceCid
		//sourceFileUrl := carFile.CarFileUrl
		//md5 := carFile.CarFileMd5
		fileSize := carFile.CarFileSize
		minerPrice, minerVerifiedPrice, _, _ := client.LotusGetMinerConfig(minerFid)

		var price float64
		if config.GetConfig().Sender.VerifiedDeal {
			if minerVerifiedPrice == nil {
				return false
			}
			price = *minerVerifiedPrice
		} else {
			if minerPrice == nil {
				return false
			}
			price = *minerPrice
		}

		maxPrice := config.GetConfig().Sender.MaxPrice
		maxPriceFloat, err := strconv.ParseFloat(maxPrice, 32)
		if err == nil {
			logs.GetLogger().Error("Failed to convert maxPrice to float.")
			return false
		}
		if price > maxPriceFloat {
			msg := fmt.Sprintf("miner %s price %f higher than max price %s", minerFid, price, maxPrice)
			logs.GetLogger().Warn(msg)
			continue
		}

		if fileSize <= 0 {
			msg := fmt.Sprintf("file %s is too small", carFile.CarFilePath)
			logs.GetLogger().Error(msg)
			continue
		}
		pieceSize, sectorSize := calculatePieceSizeFromFileSize(fileSize)
		cost := calculateRealCost(sectorSize, price)
		dealCid, startEpoch := client.LotusProposeOfflineDeal(price, cost, pieceSize, carFile.DataCid, carFile.PieceCid, minerFid)
		outputCsvPath := ""
		carFile.MinerFid = &minerFid
		carFile.DealCid = *dealCid
		carFile.StartEpoch = strconv.Itoa(*startEpoch)

		logs.GetLogger().Info("Swan deal final CSV Generated: %s", outputCsvPath)

	}

	return true
}

func createCsv4Deal(task model.Task, carFiles []*model.FileDesc, minerId *string, outDir string) error {
	csvFileName := task.TaskName + ".csv"
	csvFilePath := filepath.Join(outDir, csvFileName)

	logs.GetLogger().Info("Swan task CSV Generated: ", csvFilePath)

	headers := []string{
		"uuid",
		"miner_id",
		"file_source_url",
		"md5",
		"start_epoch",
		"deal_cid",
	}

	file, err := os.Create(csvFilePath)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.Write(headers)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	for _, carFile := range carFiles {
		var columns []string
		columns = append(columns, carFile.Uuid)
		if minerId != nil {
			columns = append(columns, *minerId)
		} else {
			columns = append(columns, "")
		}
		columns = append(columns, carFile.CarFileUrl)
		columns = append(columns, carFile.CarFileMd5)
		columns = append(columns, carFile.StartEpoch)
		columns = append(columns, carFile.DealCid)

		err = writer.Write(columns)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
	}

	if config.GetConfig().Sender.OfflineMode {
		logs.GetLogger().Info("Working in Offline Mode. You need to manually send out task on filwan.com.")
		return nil
	}

	logs.GetLogger().Info("Working in Online Mode. A swan task will be created on the filwan.com after process done. ")

	swanClient := client.SwanGetClient()

	response := swanClient.SwanCreateTask(task, csvFilePath)
	logs.GetLogger().Info(response)

	return nil
}

func createCsv4SendDeal(carFiles []*model.FileDesc, minerId *string, outDir string, task *model.Task) error {
	csvFileName := task.TaskName + ".csv"
	csvFilePath := filepath.Join(outDir, csvFileName)

	logs.GetLogger().Info("Swan task CSV Generated: ", csvFilePath)

	headers := []string{
		"uuid",
		"miner_id",
		"file_source_url",
		"md5",
		"start_epoch",
		"deal_cid",
	}

	file, err := os.Create(csvFilePath)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.Write(headers)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	for _, carFile := range carFiles {
		columns := []string{}
		columns = append(columns, carFile.Uuid)
		if minerId != nil {
			columns = append(columns, *minerId)
		} else {
			columns = append(columns, "")
		}
		columns = append(columns, carFile.CarFileUrl)
		columns = append(columns, carFile.CarFileMd5)
		columns = append(columns, carFile.StartEpoch)
		columns = append(columns, carFile.DealCid)

		err = writer.Write(columns)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
	}

	return nil
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
