package subcommand

import (
	"encoding/csv"
	"encoding/json"
	"go-swan-client/common/utils"
	"go-swan-client/logs"
	"go-swan-client/model"

	"io/ioutil"
	"os"
	"strconv"
)

func generateJsonFile(carFiles []*model.FileDesc, outputDir string) {
	jsonFilePath := utils.GetPath(outputDir, "car.json")
	content, err := json.MarshalIndent(carFiles, "", " ")
	if err != nil {
		logs.GetLogger().Fatal(err)
	}

	err = ioutil.WriteFile(jsonFilePath, content, 0644)
	if err != nil {
		logs.GetLogger().Fatal(err)
	}
}

func readCarFilesFromJsonFile(inputDir string) []*model.FileDesc {
	jsonFilePath := utils.GetPath(inputDir, "car.json")

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

func generateCsvFile(carFiles []*model.FileDesc, outputDir, csvFileName string) {
	csvPath := utils.GetPath(outputDir, csvFileName)

	var headers []string
	headers = append(headers, "car_file_name")
	headers = append(headers, "car_file_path")
	headers = append(headers, "piece_cid")
	headers = append(headers, "data_cid")
	headers = append(headers, "car_file_size")
	headers = append(headers, "car_file_md5")
	headers = append(headers, "source_file_name")
	headers = append(headers, "source_file_path")
	headers = append(headers, "source_file_size")
	headers = append(headers, "source_file_md5")
	headers = append(headers, "car_file_url")

	file, err := os.Create(csvPath)
	if err != nil {
		logs.GetLogger().Fatal(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.Write(headers)
	if err != nil {
		logs.GetLogger().Fatal(err)
	}

	for _, carFile := range carFiles {
		var columns []string
		columns = append(columns, carFile.CarFileName)
		columns = append(columns, carFile.CarFilePath)
		columns = append(columns, carFile.PieceCid)
		columns = append(columns, carFile.DataCid)
		columns = append(columns, strconv.FormatInt(carFile.CarFileSize, 10))
		columns = append(columns, carFile.CarFileMd5)
		columns = append(columns, carFile.SourceFileName)
		columns = append(columns, carFile.SourceFilePath)
		columns = append(columns, strconv.FormatInt(carFile.CarFileSize, 10))
		columns = append(columns, strconv.FormatBool(carFile.SourceFileMd5))
		columns = append(columns, carFile.CarFileUrl)

		err = writer.Write(columns)
		if err != nil {
			logs.GetLogger().Fatal(err)
		}
	}
}
