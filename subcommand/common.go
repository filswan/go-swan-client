package subcommand

import (
	"encoding/csv"
	"encoding/json"
	"go-swan-client/logs"
	"go-swan-client/model"
	"path/filepath"

	"io/ioutil"
	"os"
	"strconv"
)

const (
	JSON_FILE_NAME_AFTER_CAR    = "car.json"
	JSON_FILE_NAME_AFTER_UPLOAD = "upload.json"
	JSON_FILE_NAME_AFTER_TASK   = "task.json"
	JSON_FILE_NAME_AFTER_DEAL   = "deal.json"
)

func WriteCarFilesToJsonFile(carFiles []*model.FileDesc, outputDir, jsonFilename string) {
	jsonFilePath := filepath.Join(outputDir, jsonFilename)
	content, err := json.MarshalIndent(carFiles, "", " ")
	if err != nil {
		logs.GetLogger().Fatal(err)
	}

	err = ioutil.WriteFile(jsonFilePath, content, 0644)
	if err != nil {
		logs.GetLogger().Fatal(err)
	}
}

func ReadCarFilesFromJsonFile(inputDir, jsonFilename string) []*model.FileDesc {
	jsonFilePath := filepath.Join(inputDir, jsonFilename)

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

func GenerateMetadataCsv(minerFid *string, carFiles []*model.FileDesc, outDir, csvFileName string) bool {
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
		if minerFid != nil {
			columns = append(columns, *minerFid)
		} else {
			columns = append(columns, "")
		}
		columns = append(columns, carFile.StartEpoch)

		err = writer.Write(columns)
		if err != nil {
			logs.GetLogger().Error(err)
			return false
		}
	}

	logs.GetLogger().Info("Metadata CSV Generated: ", csvFilePath)

	return true
}
