package subcommand

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"go-swan-client/logs"

	"go-swan-client/model"

	"go-swan-client/common/utils"

	"go-swan-client/common/client"
	"go-swan-client/common/constants"

	"github.com/codingsince1985/checksum"
)

func GenerateCarFiles(inputDir string, outputDir *string) (*string, []*model.FileDesc, error) {
	err := CheckInputDir(inputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	outputDir, err = CreateOutputDir(outputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	srcFiles, err := ioutil.ReadDir(inputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	carFiles := []*model.FileDesc{}

	for _, srcFile := range srcFiles {
		carFile := model.FileDesc{}
		carFile.SourceFileName = srcFile.Name()
		carFile.SourceFilePath = filepath.Join(inputDir, carFile.SourceFileName)
		carFile.SourceFileSize = srcFile.Size()
		carFile.CarFileName = carFile.SourceFileName + ".car"
		carFile.CarFilePath = filepath.Join(*outputDir, carFile.CarFileName)

		err := client.LotusClientGenCar(carFile.SourceFilePath, carFile.CarFilePath, false)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, nil, err
		}

		pieceCid := client.LotusClientCalcCommP(carFile.CarFilePath)
		if pieceCid == nil {
			err := fmt.Errorf("failed to generate piece cid")
			logs.GetLogger().Error(err)
			return nil, nil, err
		}

		carFile.PieceCid = *pieceCid

		dataCid, err := client.LotusClientImport(carFile.CarFilePath, true)
		if err != nil {
			err := fmt.Errorf("failed to import car file")
			logs.GetLogger().Error(err)
			return nil, nil, err
		}

		carFile.DataCid = *dataCid

		carFile.CarFileSize = utils.GetFileSize(carFile.CarFilePath)

		srcFileMd5, err := checksum.MD5sum(carFile.SourceFilePath)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, nil, err
		}
		carFile.SourceFileMd5 = srcFileMd5

		carFileMd5, err := checksum.MD5sum(carFile.CarFilePath)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, nil, err
		}
		carFile.CarFileMd5 = carFileMd5

		carFiles = append(carFiles, &carFile)
	}

	err = WriteCarFilesToFiles(carFiles, *outputDir, constants.JSON_FILE_NAME_BY_CAR, constants.CSV_FILE_NAME_BY_CAR)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	logs.GetLogger().Info("Car files output dir: ", *outputDir)
	logs.GetLogger().Info("Please upload car files to web server or ipfs server.")

	return outputDir, carFiles, nil
}
