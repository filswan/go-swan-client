package subcommand

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"go-swan-client/model"

	"go-swan-client/logs"

	"go-swan-client/common/client"
	"go-swan-client/common/utils"
	"go-swan-client/config"

	"go-swan-client/common/constants"

	"io/ioutil"
	"os"
	"strconv"
)

func CheckDealConfig(dealConfig *model.ConfDeal) error {
	minerPrice, minerVerifiedPrice, _, _ := client.LotusGetMinerConfig(dealConfig.MinerFid)

	if dealConfig.SenderWallet == "" {
		err := fmt.Errorf("sender.wallet should be set in config file")
		logs.GetLogger().Error(err)
		return err
	}

	if dealConfig.VerifiedDeal {
		if minerVerifiedPrice == nil {
			err := fmt.Errorf("cannot get miner verified price for verified deal")
			logs.GetLogger().Error(err)
			return err
		}
		dealConfig.MinerPrice = *minerVerifiedPrice
		logs.GetLogger().Info("Miner price is:", *minerVerifiedPrice)
	} else {
		if minerPrice == nil {
			err := fmt.Errorf("cannot get miner price for non-verified deal")
			logs.GetLogger().Error(err)
			return err
		}
		dealConfig.MinerPrice = *minerPrice
		logs.GetLogger().Info("Miner price is:", *minerPrice)
	}

	logs.GetLogger().Info("Miner price is:", dealConfig.MinerPrice, " MaxPrice:", dealConfig.MaxPrice, " VerifiedDeal:", dealConfig.VerifiedDeal)
	priceCmp := dealConfig.MaxPrice.Cmp(dealConfig.MinerPrice)
	logs.GetLogger().Info("priceCmp:", priceCmp)
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

func CreateOutputDir(outputDir *string) (*string, error) {
	if outputDir == nil || len(*outputDir) == 0 {
		if outputDir == nil {
			outDir := filepath.Join(config.GetConfig().Sender.OutputDir, time.Now().Format("2006-01-02_15:04:05"))
			outputDir = &outDir
		} else {
			*outputDir = filepath.Join(config.GetConfig().Sender.OutputDir, time.Now().Format("2006-01-02_15:04:05"))
		}

		logs.GetLogger().Info("output-dir is not provided, use default:", *outputDir)
	}

	err := os.MkdirAll(*outputDir, os.ModePerm)
	if err != nil {
		err := fmt.Errorf("%s, failed to create output dir:%s", err.Error(), *outputDir)
		logs.GetLogger().Error(err)
		return nil, err
	}

	return outputDir, nil
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
	headers = append(headers, "source_file_url")
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
		columns = append(columns, carFile.Uuid)
		columns = append(columns, carFile.SourceFileName)
		columns = append(columns, carFile.SourceFilePath)
		columns = append(columns, carFile.SourceFileMd5)
		columns = append(columns, carFile.SourceFileUrl)
		columns = append(columns, strconv.FormatInt(carFile.SourceFileSize, 10))
		columns = append(columns, carFile.CarFileName)
		columns = append(columns, carFile.CarFilePath)
		columns = append(columns, carFile.CarFileMd5)
		columns = append(columns, carFile.CarFileUrl)
		columns = append(columns, strconv.FormatInt(carFile.CarFileSize, 10))
		columns = append(columns, carFile.DealCid)
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
		columns = append(columns, carFile.Uuid)
		if carFile.MinerFid != nil {
			columns = append(columns, *carFile.MinerFid)
		} else {
			columns = append(columns, "")
		}
		columns = append(columns, carFile.DealCid)
		columns = append(columns, carFile.DataCid)
		columns = append(columns, carFile.CarFileUrl)
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
