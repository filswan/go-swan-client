package subcommand

import (
	"encoding/csv"
	"encoding/json"
	"path/filepath"

	"go-swan-client/logs"
	"go-swan-client/model"

	"io/ioutil"
	"os"
	"strconv"
)

const (
	JSON_FILE_NAME_BY_CAR         = "car.json"
	JSON_FILE_NAME_BY_UPLOAD      = "upload.json"
	JSON_FILE_NAME_BY_TASK_SUFFIX = "task.json"
	JSON_FILE_NAME_BY_DEAL_SUFFIX = "deal.json"
	JSON_FILE_NAME_BY_AUTO_SUFFIX = "deal_autobid.json"

	CSV_FILE_NAME_BY_CAR         = "car.csv"
	CSV_FILE_NAME_BY_UPLOAD      = "upload.csv"
	CSV_FILE_NAME_BY_TASK_SUFFIX = "task.csv"
	CSV_FILE_NAME_BY_DEAL_SUFFIX = "deal.csv"
	CSV_FILE_NAME_BY_AUTO_SUFFIX = "deal_autobid.csv"
)

func WriteCarFilesToFiles(carFiles []*model.FileDesc, outputDir, jsonFilename, csvFileName string) bool {
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		logs.GetLogger().Error(err)
		logs.GetLogger().Error("Failed to create output dir:", outputDir)
		return false
	}

	result := WriteCarFilesToJsonFile(carFiles, outputDir, jsonFilename)
	if !result {
		logs.GetLogger().Error("Failed to generate json file.")
		return result
	}

	result = WriteCarFilesToCsvFile(carFiles, outputDir, csvFileName)
	if !result {
		logs.GetLogger().Error("Failed to generate json file.")
		return result
	}

	return true
}

func WriteCarFilesToJsonFile(carFiles []*model.FileDesc, outputDir, jsonFilename string) bool {
	jsonFilePath := filepath.Join(outputDir, jsonFilename)
	content, err := json.MarshalIndent(carFiles, "", " ")
	if err != nil {
		logs.GetLogger().Error(err)
		return false
	}

	err = ioutil.WriteFile(jsonFilePath, content, 0644)
	if err != nil {
		logs.GetLogger().Error(err)
		return false
	}

	logs.GetLogger().Info("Metadata CSV Generated: ", jsonFilePath)
	return true
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

func WriteCarFilesToCsvFile(carFiles []*model.FileDesc, outDir, csvFileName string) bool {
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
		return false
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.Write(headers)
	if err != nil {
		logs.GetLogger().Error(err)
		return false
	}

	for _, carFile := range carFiles {
		var columns []string
		columns = append(columns, carFile.Uuid)
		columns = append(columns, carFile.SourceFileName)
		columns = append(columns, carFile.SourceFilePath)
		columns = append(columns, strconv.FormatBool(carFile.SourceFileMd5))
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
			return false
		}
	}

	logs.GetLogger().Info("Metadata CSV Generated: ", csvFilePath)

	return true
}

func CreateCsv4TaskDeal(carFiles []*model.FileDesc, minerId *string, outDir, csvFileName string) (string, error) {
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
		if minerId != nil {
			columns = append(columns, *minerId)
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
