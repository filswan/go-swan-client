package subcommand

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"go-swan-client/common/client"
	"go-swan-client/common/constants"
	"go-swan-client/common/utils"
	"go-swan-client/config"
	"go-swan-client/logs"
	"go-swan-client/model"

	"github.com/codingsince1985/checksum"
)

func GenerateCarFiles(inputDir, outputDir *string) (*string, []*model.FileDesc, error) {
	if inputDir == nil || len(*inputDir) == 0 {
		err := fmt.Errorf("please provide input dir")
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	if utils.GetPathType(*inputDir) != constants.PATH_TYPE_DIR {
		err := fmt.Errorf("input dir: %s not exists", *inputDir)
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	outputDir, err := CreateOutputDir(outputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	srcFiles, err := ioutil.ReadDir(*inputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, nil, err
	}

	carFiles := []*model.FileDesc{}

	for _, srcFile := range srcFiles {
		carFile := model.FileDesc{}
		carFile.SourceFileName = srcFile.Name()
		carFile.SourceFilePath = filepath.Join(*inputDir, carFile.SourceFileName)
		carFile.SourceFileSize = srcFile.Size()
		carFile.CarFileName = carFile.SourceFileName + ".car"
		carFile.CarFilePath = filepath.Join(*outputDir, carFile.CarFileName)

		err := client.LotusGenerateCar(carFile.SourceFilePath, carFile.CarFilePath)
		if err != nil {
			logs.GetLogger().Fatal("Failed to generate car file.")
		}

		pieceCid := client.LotusGeneratePieceCid(carFile.CarFilePath)
		if pieceCid == nil {
			logs.GetLogger().Fatal("Failed to generate piece cid.")
		}

		carFile.PieceCid = *pieceCid

		dataCid := client.LotusImportCarFile(carFile.CarFilePath)
		if dataCid == nil {
			logs.GetLogger().Fatal("Failed to import car file.")
		}

		carFile.DataCid = *dataCid

		carFile.CarFileSize = utils.GetFileSize(carFile.CarFilePath)

		if config.GetConfig().Sender.GenerateMd5 {
			carFileMd5, err := checksum.MD5sum(carFile.CarFilePath)
			if err != nil {
				logs.GetLogger().Error(err)
				logs.GetLogger().Fatal("Failed to generate md5 for car file:", carFile.CarFilePath)
			}
			logs.GetLogger().Info("carFileMd5:", carFileMd5)
			carFile.CarFileMd5 = carFileMd5
		}

		carFiles = append(carFiles, &carFile)
	}

	WriteCarFilesToFiles(carFiles, *outputDir, constants.JSON_FILE_NAME_BY_CAR, constants.CSV_FILE_NAME_BY_CAR)
	logs.GetLogger().Info("Car files output dir: ", outputDir)
	logs.GetLogger().Info("Please upload car files to web server or ipfs server.")

	return outputDir, carFiles, nil
}
