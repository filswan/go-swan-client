package subcommand

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/filswan/go-swan-client/model"
	"github.com/shopspring/decimal"

	"github.com/filswan/go-swan-lib/client/lotus"
	"github.com/filswan/go-swan-lib/constants"
	"github.com/filswan/go-swan-lib/logs"
	libmodel "github.com/filswan/go-swan-lib/model"
	"github.com/filswan/go-swan-lib/utils"

	"io/ioutil"
	"os"
	"strconv"
)

const (
	DURATION     = 1512000
	DURATION_MIN = 518400
	DURATION_MAX = 1540000
)

func CheckDuration(duration int, startEpoch int, relativeEpochFromMainNetwork int) error {
	if duration == 0 {
		return nil
	}

	if duration < DURATION_MIN || duration > DURATION_MAX {
		err := fmt.Errorf("deal duration out of bounds (min, max, provided): %d, %d, %d", DURATION_MIN, DURATION_MAX, duration)
		logs.GetLogger().Error(err)
		return err
	}

	currentEpoch := utils.GetCurrentEpoch() + relativeEpochFromMainNetwork
	endEpoch := startEpoch + duration

	epoch2EndfromNow := endEpoch - currentEpoch
	if epoch2EndfromNow >= DURATION_MAX {
		err := fmt.Errorf("invalid deal end epoch %d: cannot be more than %d past current epoch %d", endEpoch, DURATION_MAX, currentEpoch)
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}

func GetDealCost(pricePerEpoch decimal.Decimal, duration int) string {
	durationDecimal := decimal.NewFromInt(int64(duration))
	cost := pricePerEpoch.Mul(durationDecimal)
	cost = cost.Mul(decimal.NewFromFloat(constants.LOTUS_PRICE_MULTIPLE))

	return cost.String()
}

func IsTaskSourceRight(confDeal *model.ConfDeal, task libmodel.Task) bool {
	if confDeal == nil {
		return false
	}

	if confDeal.DealSourceIds == nil || len(confDeal.DealSourceIds) == 0 {
		return false
	}

	for _, sourceId := range confDeal.DealSourceIds {
		if task.SourceId == sourceId {
			return true
		}
	}

	return false
}

func GetDefaultTaskName() string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	randStr := utils.RandStringRunes(letterRunes, 6)
	taskName := "swan-task-" + randStr
	return taskName
}

func CheckDealConfig(confDeal *model.ConfDeal) error {
	if confDeal == nil {
		err := fmt.Errorf("parameter confDeal is nil")
		logs.GetLogger().Error(err)
		return err
	}

	lotusClient, err := lotus.LotusGetClient(confDeal.LotusClientApiUrl, confDeal.LotusClientAccessToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	minerPrice, minerVerifiedPrice, _, _ := lotusClient.LotusGetMinerConfig(confDeal.MinerFid)

	if confDeal.SenderWallet == "" {
		err := fmt.Errorf("wallet should be set")
		logs.GetLogger().Error(err)
		return err
	}

	if confDeal.VerifiedDeal {
		if minerVerifiedPrice == nil {
			err := fmt.Errorf("miner:%s,cannot get miner verified price for verified deal", confDeal.MinerFid)
			logs.GetLogger().Error(err)
			return err
		}
		confDeal.MinerPrice = *minerVerifiedPrice
		logs.GetLogger().Info("miner:", confDeal.MinerFid, ",price is:", *minerVerifiedPrice)
	} else {
		if minerPrice == nil {
			err := fmt.Errorf("miner:%s,cannot get miner price for non-verified deal", confDeal.MinerFid)
			logs.GetLogger().Error(err)
			return err
		}
		confDeal.MinerPrice = *minerPrice
		logs.GetLogger().Info("miner:", confDeal.MinerFid, ",price is:", *minerPrice)
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

	if confDeal.Duration == 0 {
		confDeal.Duration = DURATION
	}

	err = CheckDuration(confDeal.Duration, confDeal.StartEpoch, confDeal.RelativeEpochFromMainNetwork)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

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

func WriteCarFilesToFiles(carFiles []*libmodel.FileDesc, outputDir, jsonFilename, csvFileName string) (*string, error) {
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	jsonFilePath, err := WriteCarFilesToJsonFile(carFiles, outputDir, jsonFilename)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	err = WriteCarFilesToCsvFile(carFiles, outputDir, csvFileName)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return jsonFilePath, nil
}

func WriteCarFilesToJsonFile(carFiles []*libmodel.FileDesc, outputDir, jsonFilename string) (*string, error) {
	jsonFilePath := filepath.Join(outputDir, jsonFilename)
	content, err := json.MarshalIndent(carFiles, "", " ")
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	err = ioutil.WriteFile(jsonFilePath, content, 0644)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	logs.GetLogger().Info("Metadata json generated: ", jsonFilePath)
	return &jsonFilePath, nil
}

func ReadCarFilesFromJsonFile(inputDir, jsonFilename string) []*libmodel.FileDesc {
	jsonFilePath := filepath.Join(inputDir, jsonFilename)
	result := ReadCarFilesFromJsonFileByFullPath(jsonFilePath)
	return result
}

func ReadCarFilesFromJsonFileByFullPath(jsonFilePath string) []*libmodel.FileDesc {
	contents, err := ioutil.ReadFile(jsonFilePath)
	if err != nil {
		logs.GetLogger().Error("Failed to read: ", jsonFilePath)
		return nil
	}

	carFiles := []*libmodel.FileDesc{}

	err = json.Unmarshal(contents, &carFiles)
	if err != nil {
		logs.GetLogger().Error("Failed to read: ", jsonFilePath)
		return nil
	}

	return carFiles
}

func WriteCarFilesToCsvFile(carFiles []*libmodel.FileDesc, outDir, csvFileName string) error {
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
	headers = append(headers, "source_id")
	headers = append(headers, "cost")

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
		columns = append(columns, carFile.SourceFileName)
		columns = append(columns, carFile.SourceFilePath)
		columns = append(columns, carFile.SourceFileMd5)
		columns = append(columns, strconv.FormatInt(carFile.SourceFileSize, 10))
		columns = append(columns, carFile.CarFileName)
		columns = append(columns, carFile.CarFilePath)
		columns = append(columns, carFile.CarFileMd5)
		columns = append(columns, carFile.CarFileUrl)
		columns = append(columns, strconv.FormatInt(carFile.CarFileSize, 10))
		columns = append(columns, carFile.DealCid)
		columns = append(columns, carFile.DataCid)
		columns = append(columns, carFile.PieceCid)
		columns = append(columns, carFile.MinerFid)

		if carFile.StartEpoch != nil {
			columns = append(columns, strconv.Itoa(*carFile.StartEpoch))
		} else {
			columns = append(columns, "")
		}

		if carFile.SourceId != nil {
			columns = append(columns, strconv.Itoa(*carFile.SourceId))
		} else {
			columns = append(columns, "")
		}
		columns = append(columns, carFile.Cost)

		err = writer.Write(columns)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
	}

	logs.GetLogger().Info("Metadata csv generated: ", csvFilePath)

	return nil
}

func CreateCsv4TaskDeal(carFiles []*libmodel.FileDesc, outDir, csvFileName string) (string, error) {
	csvFilePath := filepath.Join(outDir, csvFileName)

	logs.GetLogger().Info("Swan task CSV Generated: ", csvFilePath)

	headers := []string{
		"uuid",
		"source_file_name",
		"miner_id",
		"deal_cid",
		"payload_cid",
		"file_source_url",
		"md5",
		"start_epoch",
		"piece_cid",
		"file_size",
		"cost",
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
		columns = append(columns, carFile.Uuid)
		columns = append(columns, carFile.SourceFileName)
		columns = append(columns, carFile.MinerFid)
		columns = append(columns, carFile.DealCid)
		columns = append(columns, carFile.DataCid)
		columns = append(columns, carFile.CarFileUrl)
		columns = append(columns, carFile.CarFileMd5)

		if carFile.StartEpoch != nil {
			columns = append(columns, strconv.Itoa(*carFile.StartEpoch))
		} else {
			columns = append(columns, "")
		}

		columns = append(columns, carFile.PieceCid)
		columns = append(columns, strconv.FormatInt(carFile.CarFileSize, 10))
		columns = append(columns, carFile.Cost)

		err = writer.Write(columns)
		if err != nil {
			logs.GetLogger().Error(err)
			return "", err
		}
	}

	return csvFilePath, nil
}
