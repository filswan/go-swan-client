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
)

const DURATION = "1051200"
const EPOCH_PER_HOUR = 120

type DealConfig struct {
	MinerId            string `json:"miner_id"`
	SenderWallet       string `json:"sender_wallet"`
	MaxPrice           string `json:"max_price"`
	VerifiedDeal       bool   `json:"verified_deal"`
	FastRetrieval      bool   `json:"fast_retrieval"`
	EpochIntervalHours int    `json:"epoch_interval_hours"`
	SkipConfirmation   bool   `json:"skip_confirmation"`
}

func sendDeals(outputDir *string, task model.Task, carFiles []*model.FileDesc, taskUuid string) {
	dealConfig := DealConfig{
		MinerId:            *task.MinerId,
		SenderWallet:       config.GetConfig().Sender.Wallet,
		MaxPrice:           config.GetConfig().Sender.MaxPrice,
		VerifiedDeal:       config.GetConfig().Sender.VerifiedDeal,
		FastRetrieval:      config.GetConfig().Sender.FastRetrieval,
		EpochIntervalHours: config.GetConfig().Sender.StartEpochHours,
		SkipConfirmation:   config.GetConfig().Sender.SkipConfirmation,
	}

	if outputDir == nil {
		outDir := config.GetConfig().Sender.OutputDir
		outputDir = &outDir
	}

	sendDeals2Miner(dealConfig, task, *outputDir, carFiles, taskUuid)
}

func sendDeals2Miner(dealConfig DealConfig, task model.Task, outputDir string, carFiles []*model.FileDesc, taskUuid string) {
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}

	//skipConfirmation := config.GetConfig().Sender.SkipConfirmation

	minerId := ""

	err = createCsv4SendDeal(carFiles, &minerId, outputDir, nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}
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

func sendDeals2Miner1(outputDir string, taskName string, taskUuid string, minerId string, carFiles []*model.FileDesc) {
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}

	for _, carFile := range carFiles {
		//dataCid := carFile.DataCid
		//pieceCid := carFile.PieceCid
		//sourceFileUrl := carFile.CarFileUrl
		//md5 := carFile.CarFileMd5
		fileSize := carFile.CarFileSize
		minerPrice, minerVerifiedPrice, _, _ := client.LotusGetMinerConfig(minerId)

		var price float64
		if config.GetConfig().Sender.VerifiedDeal {
			if minerVerifiedPrice == nil {
				return
			}
			price = *minerVerifiedPrice
		} else {
			if minerPrice == nil {
				return
			}
			price = *minerPrice
		}

		maxPrice := config.GetConfig().Sender.MaxPrice
		maxPriceFloat, err := strconv.ParseFloat(maxPrice, 32)
		if err == nil {
			logs.GetLogger().Error("Failed to convert maxPrice to float.")
			return
		}
		if price > maxPriceFloat {
			msg := fmt.Sprintf("miner %s price %s higher than max price %s", minerId, price, maxPrice)
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
		dealCid, startEpoch := client.LotusProposeOfflineDeal(price, cost, pieceSize, carFile.DataCid, carFile.PieceCid, minerId)
		outputCsvPath := ""
		carFile.MinerId = minerId
		carFile.DealCid = *dealCid
		carFile.StartEpoch = strconv.Itoa(*startEpoch)

		logs.GetLogger().Info("Swan deal final CSV Generated: %s", outputCsvPath)

	}
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
