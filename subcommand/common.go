package subcommand

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"

	"github.com/filswan/go-swan-client/common/client"
	"github.com/filswan/go-swan-client/common/constants"
	"github.com/filswan/go-swan-client/common/utils"
	"github.com/filswan/go-swan-client/logs"
	"github.com/filswan/go-swan-client/model"

	"io/ioutil"
	"os"
	"strconv"

	"github.com/shopspring/decimal"
)

func CheckDealConfig(confDeal *model.ConfDeal) error {
	minerPrice, minerVerifiedPrice, _, _ := client.LotusGetMinerConfig(*confDeal.MinerFid)

	if confDeal.SenderWallet == "" {
		err := fmt.Errorf("wallet should be set")
		logs.GetLogger().Error(err)
		return err
	}

	if confDeal.VerifiedDeal {
		if minerVerifiedPrice == nil {
			err := fmt.Errorf("cannot get miner verified price for verified deal")
			logs.GetLogger().Error(err)
			return err
		}
		confDeal.MinerPrice = *minerVerifiedPrice
		logs.GetLogger().Info("Miner price is:", *minerVerifiedPrice)
	} else {
		if minerPrice == nil {
			err := fmt.Errorf("cannot get miner price for non-verified deal")
			logs.GetLogger().Error(err)
			return err
		}
		confDeal.MinerPrice = *minerPrice
		logs.GetLogger().Info("Miner price is:", *minerPrice)
	}

	logs.GetLogger().Info("Miner price is:", confDeal.MinerPrice, " MaxPrice:", confDeal.MaxPrice, " VerifiedDeal:", confDeal.VerifiedDeal)
	priceCmp := confDeal.MaxPrice.Cmp(confDeal.MinerPrice)
	//logs.GetLogger().Info("priceCmp:", priceCmp)
	if priceCmp < 0 {
		err := fmt.Errorf("miner price is higher than deal max price")
		logs.GetLogger().Error(err)
		return err
	}

	logs.GetLogger().Info("Deal check passed.")

	return nil
}

func CheckInputDir(inputDir string) error {
	if len(inputDir) == 0 {
		err := fmt.Errorf("please provide -input-dir")
		logs.GetLogger().Error(err)
		return err
	}

	if utils.GetPathType(inputDir) != constants.PATH_TYPE_DIR {
		err := fmt.Errorf("%s is not a directory", inputDir)
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}

func CreateOutputDir(outputDir string) error {
	if len(outputDir) == 0 {
		err := fmt.Errorf("output dir is not provided")
		logs.GetLogger().Info(err)
		return err
	}

	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		err := fmt.Errorf("%s, failed to create output dir:%s", err.Error(), outputDir)
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}

func WriteCarFilesToFiles(carFiles []*model.FileDesc, outputDir, jsonFilename, csvFileName string) error {
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	err = WriteCarFilesToJsonFile(carFiles, outputDir, jsonFilename)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	err = WriteCarFilesToCsvFile(carFiles, outputDir, csvFileName)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}

func WriteCarFilesToJsonFile(carFiles []*model.FileDesc, outputDir, jsonFilename string) error {
	jsonFilePath := filepath.Join(outputDir, jsonFilename)
	content, err := json.MarshalIndent(carFiles, "", " ")
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	err = ioutil.WriteFile(jsonFilePath, content, 0644)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	logs.GetLogger().Info("Metadata json generated: ", jsonFilePath)
	return nil
}

func ReadCarFilesFromJsonFile(inputDir, jsonFilename string) []*model.FileDesc {
	jsonFilePath := filepath.Join(inputDir, jsonFilename)
	result := ReadCarFilesFromJsonFileByFullPath(jsonFilePath)
	return result
}

func ReadCarFilesFromJsonFileByFullPath(jsonFilePath string) []*model.FileDesc {
	contents, err := ioutil.ReadFile(jsonFilePath)
	if err != nil {
		logs.GetLogger().Error("Failed to read: ", jsonFilePath)
		return nil
	}

	carFiles := []*model.FileDesc{}

	err = json.Unmarshal(contents, &carFiles)
	if err != nil {
		logs.GetLogger().Error("Failed to read: ", jsonFilePath)
		return nil
	}

	return carFiles
}

func WriteCarFilesToCsvFile(carFiles []*model.FileDesc, outDir, csvFileName string) error {
	csvFilePath := filepath.Join(outDir, csvFileName)
	var headers []string
	headers = append(headers, "uuid")
	headers = append(headers, "source_file_name")
	headers = append(headers, "source_file_path")
	headers = append(headers, "source_file_md5")
	headers = append(headers, "source_file_size")
	headers = append(headers, "car_file_name")
	headers = append(headers, "car_file_path")
	headers = append(headers, "car_file_md5")
	headers = append(headers, "car_file_url")
	headers = append(headers, "car_file_size")
	headers = append(headers, "deal_cid")
	headers = append(headers, "data_cid")
	headers = append(headers, "piece_cid")
	headers = append(headers, "miner_id")
	headers = append(headers, "start_epoch")

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
		if carFile.Uuid != nil {
			columns = append(columns, *carFile.Uuid)
		} else {
			columns = append(columns, "")
		}

		columns = append(columns, carFile.SourceFileName)
		columns = append(columns, carFile.SourceFilePath)
		columns = append(columns, carFile.SourceFileMd5)
		columns = append(columns, strconv.FormatInt(carFile.SourceFileSize, 10))
		columns = append(columns, carFile.CarFileName)
		columns = append(columns, carFile.CarFilePath)
		columns = append(columns, carFile.CarFileMd5)

		if carFile.CarFileUrl != nil {
			columns = append(columns, *carFile.CarFileUrl)
		} else {
			columns = append(columns, "")
		}

		columns = append(columns, strconv.FormatInt(carFile.CarFileSize, 10))
		if carFile.DealCid != nil {
			columns = append(columns, *carFile.DealCid)
		} else {
			columns = append(columns, "")
		}

		columns = append(columns, carFile.DataCid)
		columns = append(columns, carFile.PieceCid)
		if carFile.MinerFid != nil {
			columns = append(columns, *carFile.MinerFid)
		} else {
			columns = append(columns, "")
		}
		columns = append(columns, strconv.Itoa(carFile.StartEpoch))

		err = writer.Write(columns)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
	}

	logs.GetLogger().Info("Metadata csv generated: ", csvFilePath)

	return nil
}

func CreateCsv4TaskDeal(carFiles []*model.FileDesc, outDir, csvFileName string) (string, error) {
	csvFilePath := filepath.Join(outDir, csvFileName)

	logs.GetLogger().Info("Swan task CSV Generated: ", csvFilePath)

	headers := []string{
		"uuid",
		"miner_id",
		"deal_cid",
		"payload_cid",
		"file_source_url",
		"md5",
		"start_epoch",
		"piece_cid",
		"file_size",
	}

	file, err := os.Create(csvFilePath)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.Write(headers)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}

	for _, carFile := range carFiles {
		var columns []string
		if carFile.Uuid != nil {
			columns = append(columns, *carFile.Uuid)
		} else {
			columns = append(columns, "")
		}

		if carFile.MinerFid != nil {
			columns = append(columns, *carFile.MinerFid)
		} else {
			columns = append(columns, "")
		}
		if carFile.DealCid != nil {
			columns = append(columns, *carFile.DealCid)
		} else {
			columns = append(columns, "")
		}

		columns = append(columns, carFile.DataCid)

		if carFile.CarFileUrl != nil {
			columns = append(columns, *carFile.CarFileUrl)
		} else {
			columns = append(columns, "")
		}

		columns = append(columns, carFile.CarFileMd5)
		columns = append(columns, strconv.Itoa(carFile.StartEpoch))
		columns = append(columns, carFile.PieceCid)
		columns = append(columns, strconv.FormatInt(carFile.CarFileSize, 10))

		err = writer.Write(columns)
		if err != nil {
			logs.GetLogger().Error(err)
			return "", err
		}
	}

	return csvFilePath, nil
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
