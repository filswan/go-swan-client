package command

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/filswan/go-swan-lib/logs"
	libmodel "github.com/filswan/go-swan-lib/model"
	"github.com/filswan/go-swan-lib/utils"

	"io/ioutil"
	"os"
)

const (
	CMD_CAR     = "car"
	CMD_GOCAR   = "gocar"
	CMD_IPFSCAR = "ipfscar"
	CMD_UPLOAD  = "upload"
	CMD_TASK    = "task"
	CMD_DEAL    = "deal"
	CMD_AUTO    = "auto"

	JSON_FILE_NAME_CAR_UPLOAD = "car.json"
	JSON_FILE_NAME_TASK       = "-metadata.json"
	JSON_FILE_NAME_DEAL       = "-deals.json"
	JSON_FILE_NAME_DEAL_AUTO  = "-auto-deals.json"
)

func checkInputDir(inputDir string) error {
	if utils.IsStrEmpty(&inputDir) {
		err := fmt.Errorf("input directory is required")
		logs.GetLogger().Error(err)
		return err
	}

	if !utils.IsDirExists(inputDir) {
		err := fmt.Errorf("directory:%s not exists", inputDir)
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}

func createOutputDir(outputDir string) error {
	if utils.IsStrEmpty(&outputDir) {
		err := fmt.Errorf("directory path is required")
		logs.GetLogger().Info(err)
		return err
	}

	if utils.IsDirExists(outputDir) {
		return nil
	}

	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		err := fmt.Errorf("failed to create dir:%s,%s", outputDir, err.Error())
		logs.GetLogger().Error(err)
		return err
	}

	logs.GetLogger().Info(outputDir, " created")
	return nil
}

func WriteFileDescsToJsonFile(carFiles []*libmodel.FileDesc, outputDir, jsonFileName string) (*string, error) {
	jsonFilePath := filepath.Join(outputDir, jsonFileName)
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

	logs.GetLogger().Info("Metadata json file generated: ", jsonFilePath)
	return &jsonFilePath, nil
}

func ReadFileDescsFromJsonFile(inputDir, jsonFilename string) []*libmodel.FileDesc {
	jsonFilePath := filepath.Join(inputDir, jsonFilename)
	result := ReadFileDescsFromJsonFileByFullPath(jsonFilePath)
	return result
}

func ReadFileDescsFromJsonFileByFullPath(jsonFilePath string) []*libmodel.FileDesc {
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

func GetDeals(carFiles []*libmodel.FileDesc) ([]*Deal, error) {
	deals := []*Deal{}
	for _, carFile := range carFiles {
		deal := Deal{
			Uuid:           carFile.Uuid,
			SourceFileName: carFile.SourceFileName,
			//MinerId:        carFile.MinerFid,
			//DealCid:        carFile.DealCid,
			PayloadCid:    carFile.PayloadCid,
			FileSourceUrl: carFile.CarFileUrl,
			Md5:           carFile.CarFileMd5,
			StartEpoch:    carFile.StartEpoch,
			PieceCid:      carFile.PieceCid,
			FileSize:      carFile.CarFileSize,
		}
		deals = append(deals, &deal)
	}

	return deals, nil
}

type Deal struct {
	Uuid           string `json:"uuid"`
	SourceFileName string `json:"source_file_name"`
	MinerId        string `json:"miner_id"`
	DealCid        string `json:"deal_cid"`
	PayloadCid     string `json:"payload_cid"`
	FileSourceUrl  string `json:"file_source_url"`
	Md5            string `json:"md5"`
	StartEpoch     *int64 `json:"start_epoch"`
	PieceCid       string `json:"piece_cid"`
	FileSize       int64  `json:"file_size"`
	Cost           string `json:"cost"`
}
