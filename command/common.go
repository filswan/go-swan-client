package command

import (
	"encoding/json"
	"path/filepath"

	"github.com/filswan/go-swan-lib/logs"
	libmodel "github.com/filswan/go-swan-lib/model"

	"io/ioutil"
)

const (
	CMD_CAR        = "car"
	CMD_GOCAR      = "gocar"
	CMD_IPFSCAR    = "ipfscar"
	CMD_IPFSCMDCAR = "ipfscmdcar"
	CMD_UPLOAD     = "upload"
	CMD_TASK       = "task"
	CMD_DEAL       = "deal"
	CMD_AUTO       = "auto"

	JSON_FILE_NAME_CAR_UPLOAD = "car.json"
	JSON_FILE_NAME_TASK       = "-metadata.json"
	JSON_FILE_NAME_DEAL       = "-deals.json"
	JSON_FILE_NAME_DEAL_AUTO  = "-auto-deals.json"

	DIR_NAME_INPUT  = "input"
	DIR_NAME_OUTPUT = "output"
)

func WriteFileDescsToJsonFile(fileDescs []*libmodel.FileDesc, outputDir, jsonFileName string) (*string, error) {
	jsonFilePath := filepath.Join(outputDir, jsonFileName)
	content, err := json.MarshalIndent(fileDescs, "", " ")
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

func ReadFileDescsFromJsonFile(inputDir, jsonFilename string) ([]*libmodel.FileDesc, error) {
	jsonFilePath := filepath.Join(inputDir, jsonFilename)
	fileDescs, err := ReadFileDescsFromJsonFileByFullPath(jsonFilePath)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return fileDescs, nil
}

func ReadFileDescsFromJsonFileByFullPath(jsonFilePath string) ([]*libmodel.FileDesc, error) {
	contents, err := ioutil.ReadFile(jsonFilePath)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	fileDescs := []*libmodel.FileDesc{}
	err = json.Unmarshal(contents, &fileDescs)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return fileDescs, nil
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
