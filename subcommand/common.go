package subcommand

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/filswan/go-swan-client/model"

	libconstants "github.com/filswan/go-swan-lib/constants"
	"github.com/filswan/go-swan-lib/logs"
	libmodel "github.com/filswan/go-swan-lib/model"
	"github.com/filswan/go-swan-lib/utils"

	"io/ioutil"
	"os"
)

const (
	DURATION     = 1512000
	DURATION_MIN = 518400
	DURATION_MAX = 1540000

	SUBCOMMAND_CAR     = "car"
	SUBCOMMAND_GOCAR   = "gocar"
	SUBCOMMAND_IPFSCAR = "ipfscar"
	SUBCOMMAND_UPLOAD  = "upload"
	SUBCOMMAND_TASK    = "task"
	SUBCOMMAND_DEAL    = "deal"
	SUBCOMMAND_AUTO    = "auto"
)

func CheckDuration(duration int, startEpoch, relativeEpochFromMainNetwork int64) error {
	if duration < DURATION_MIN || duration > DURATION_MAX {
		err := fmt.Errorf("deal duration out of bounds (min, max, provided): %d, %d, %d", DURATION_MIN, DURATION_MAX, duration)
		logs.GetLogger().Error(err)
		return err
	}

	currentEpoch := int64(utils.GetCurrentEpoch()) + relativeEpochFromMainNetwork
	endEpoch := startEpoch + (int64)(duration)

	epoch2EndfromNow := endEpoch - currentEpoch
	if epoch2EndfromNow >= DURATION_MAX {
		err := fmt.Errorf("invalid deal end epoch %d: cannot be more than %d past current epoch %d", endEpoch, DURATION_MAX, currentEpoch)
		logs.GetLogger().Error(err)
		return err
	}

	return nil
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

func CheckInputDir(inputDir string) error {
	if len(inputDir) == 0 {
		err := fmt.Errorf("please provide -input-dir")
		logs.GetLogger().Error(err)
		return err
	}

	if utils.GetPathType(inputDir) != libconstants.PATH_TYPE_DIR {
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

	if utils.IsDirExists(outputDir) {
		return nil
	}

	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		err := fmt.Errorf("%s, failed to create output dir:%s", err.Error(), outputDir)
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
