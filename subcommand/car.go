package subcommand

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/filswan/go-swan-client/model"

	"github.com/filswan/go-swan-lib/client/lotus"
	"github.com/filswan/go-swan-lib/logs"

	"github.com/filswan/go-swan-lib/utils"

	"github.com/codingsince1985/checksum"
	"github.com/filswan/go-swan-client/common/constants"
	libmodel "github.com/filswan/go-swan-lib/model"
)

func CreateCarFiles(confCar *model.ConfCar) ([]*libmodel.FileDesc, error) {
	if confCar == nil {
		err := fmt.Errorf("parameter confCar is nil")
		logs.GetLogger().Error(err)
		return nil, err
	}

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

	carFiles := []*libmodel.FileDesc{}

	lotusClient, err := lotus.LotusGetClient(confCar.LotusClientApiUrl, confCar.LotusClientAccessToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	for _, srcFile := range srcFiles {
		carFile := libmodel.FileDesc{}
		carFile.SourceFileName = srcFile.Name()
		carFile.SourceFilePath = filepath.Join(confCar.InputDir, carFile.SourceFileName)
		carFile.SourceFileSize = srcFile.Size()
		carFile.CarFileName = carFile.SourceFileName + ".car"
		carFile.CarFilePath = filepath.Join(confCar.OutputDir, carFile.CarFileName)
		logs.GetLogger().Info("Creating car file ", carFile.CarFilePath, " for ", carFile.SourceFilePath)

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

		carFile.PayloadCid = *dataCid

		carFile.CarFileSize = utils.GetFileSize(carFile.CarFilePath)

		if confCar.GenerateMd5 {
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
		}

		carFiles = append(carFiles, &carFile)
		logs.GetLogger().Info("Car file ", carFile.CarFilePath, " created")
	}

	logs.GetLogger().Info(len(carFiles), " car files have been created to directory:", confCar.OutputDir)

	_, err = WriteCarFilesToJsonFile(carFiles, confCar.OutputDir, constants.JSON_FILE_NAME_CAR_UPLOAD, SUBCOMMAND_CAR)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	logs.GetLogger().Info("Please upload car files to web server or ipfs server.")

	return carFiles, nil
}
