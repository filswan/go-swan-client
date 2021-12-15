package subcommand

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/filswan/go-swan-client/model"
	"github.com/shopspring/decimal"

	"github.com/filswan/go-swan-lib/client/lotus"
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
	cost = cost.Mul(decimal.NewFromFloat(libconstants.LOTUS_PRICE_MULTIPLE_1E18))

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
	}

	priceCmp := confDeal.MaxPrice.Cmp(confDeal.MinerPrice)
	if priceCmp < 0 {
		logs.GetLogger().Info("Miner price is:", confDeal.MinerPrice, " MaxPrice:", confDeal.MaxPrice, " VerifiedDeal:", confDeal.VerifiedDeal)
		err := fmt.Errorf("miner price is higher than deal max price")
		logs.GetLogger().Error(err)
		return err
	}

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

func WriteCarFilesToJsonFile(carFiles []*libmodel.FileDesc, outputDir, jsonFileName string) (*string, error) {
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
	StartEpoch     *int   `json:"start_epoch"`
	PieceCid       string `json:"piece_cid"`
	FileSize       int64  `json:"file_size"`
	Cost           string `json:"cost"`
}
