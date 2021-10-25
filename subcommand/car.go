package subcommand

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/DoraNebula/go-swan-client/logs"

	"github.com/DoraNebula/go-swan-client/model"

	"github.com/DoraNebula/go-swan-client/common/client"
	"github.com/DoraNebula/go-swan-client/common/utils"

	"github.com/DoraNebula/go-swan-client/common/constants"

	"github.com/codingsince1985/checksum"
)

func GenerateCarFiles(confCar *model.ConfCar) ([]*model.FileDesc, error) {
	err := CheckInputDir(confCar.InputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	err = CreateOutputDir(confCar.OutputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	srcFiles, err := ioutil.ReadDir(confCar.InputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	carFiles := []*model.FileDesc{}

	lotusClient, err := client.LotusGetClient(confCar.LotusApiUrl, confCar.LotusAccessToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	for _, srcFile := range srcFiles {
		carFile := model.FileDesc{}
		carFile.SourceFileName = srcFile.Name()
		carFile.SourceFilePath = filepath.Join(confCar.InputDir, carFile.SourceFileName)
		carFile.SourceFileSize = srcFile.Size()
		carFile.CarFileName = carFile.SourceFileName + ".car"
		carFile.CarFilePath = filepath.Join(confCar.OutputDir, carFile.CarFileName)

		err := lotusClient.LotusClientGenCar(carFile.SourceFilePath, carFile.CarFilePath, false)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}

		pieceCid := lotusClient.LotusClientCalcCommP(carFile.CarFilePath)
		if pieceCid == nil {
			err := fmt.Errorf("failed to generate piece cid")
			logs.GetLogger().Error(err)
			return nil, err
		}

		carFile.PieceCid = *pieceCid

		dataCid, err := lotusClient.LotusClientImport(carFile.CarFilePath, true)
		if err != nil {
			err := fmt.Errorf("failed to import car file")
			logs.GetLogger().Error(err)
			return nil, err
		}

		if dataCid == nil {
			err := fmt.Errorf("failed to generate data cid for: %s", carFile.CarFilePath)
			logs.GetLogger().Error(err)
			return nil, err
		}

		carFile.DataCid = *dataCid

		carFile.CarFileSize = utils.GetFileSize(carFile.CarFilePath)

		srcFileMd5, err := checksum.MD5sum(carFile.SourceFilePath)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}
		carFile.SourceFileMd5 = srcFileMd5

		carFileMd5, err := checksum.MD5sum(carFile.CarFilePath)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}
		carFile.CarFileMd5 = carFileMd5

		carFiles = append(carFiles, &carFile)
	}

	err = WriteCarFilesToFiles(carFiles, confCar.OutputDir, constants.JSON_FILE_NAME_BY_CAR, constants.CSV_FILE_NAME_BY_CAR)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	logs.GetLogger().Info(len(carFiles), " car files have been created to directory:", confCar.OutputDir)
	logs.GetLogger().Info("Please upload car files to web server or ipfs server.")

	return carFiles, nil
}
