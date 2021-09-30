package operation

import (
	"encoding/csv"
	"encoding/json"
	"go-swan-client/common/utils"
	"go-swan-client/logs"
	"io/ioutil"
	"os"
)

func GenerateJsonFile(carFiles []*FileDesc, outputDir string) error {
	jsonFilePath := utils.GetDir(outputDir, "car.json")
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

	return nil
}

func GenerateCsvFile(carFiles []*FileDesc, outputDir string) error {
	csvPath := utils.GetDir(outputDir, "car.csv")

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
		columns = append(columns, carFile.CarFileName)
		columns = append(columns, carFile.CarFilePath)
		columns = append(columns, carFile.pieceCid)
		columns = append(columns, carFile.DataCid)
		columns = append(columns, carFile.CarFileSize)
		columns = append(columns, carFile.CarFileMd5)
		columns = append(columns, carFile.SourceFileName)
		columns = append(columns, carFile.SourceFilePath)
		columns = append(columns, carFile.CarFileSize)
		columns = append(columns, carFile.SourceFileMd5)
		columns = append(columns, "")

		err = writer.Write(columns)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
	}

	logs.GetLogger().Info("Car files output dir: ", outputDir)
	logs.GetLogger().Info("Please upload car files to web server or ipfs server.")

	return nil
}