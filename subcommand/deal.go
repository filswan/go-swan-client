package subcommand

import (
	"encoding/csv"
	"fmt"
	"go-swan-client/common/client"
	"go-swan-client/common/utils"
	"go-swan-client/config"
	"go-swan-client/logs"
	"go-swan-client/model"
	"math"
	"os"
	"strconv"
)

const DURATION = "1051200"
const EPOCH_PER_HOUR = 120

/*
class DealConfig:
    miner_id = None
    sender_wallet = None
    max_price = None
    verified_deal = None
    fast_retrieval = None
    epoch_interval_hours = None

    def __init__(self, miner_id, sender_wallet, max_price, verified_deal, fast_retrieval, epoch_interval_hours):
        self.miner_id = miner_id
        self.sender_wallet = sender_wallet
        self.max_price = max_price
        self.verified_deal = verified_deal
        self.fast_retrieval = fast_retrieval
        self.epoch_interval_hours = epoch_interval_hours
*/

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
	csvFilePath := utils.GetPath(outDir, csvFileName)

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
