package subcommand

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/codingsince1985/checksum"
	"github.com/filswan/go-swan-client/model"
	"github.com/filswan/go-swan-lib/client/ipfs"
	"github.com/filswan/go-swan-lib/client/lotus"
	"github.com/filswan/go-swan-lib/constants"
	"github.com/filswan/go-swan-lib/logs"
	libmodel "github.com/filswan/go-swan-lib/model"
	"github.com/filswan/go-swan-lib/utils"
)

func CreateIpfsCarFiles(confCar *model.ConfCar) ([]*libmodel.FileDesc, error) {
	ipfsUrlPrefix := "http://192.168.88.41:5001"
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

	lotusClient, err := lotus.LotusGetClient(confCar.LotusClientApiUrl, confCar.LotusClientAccessToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	srcFileCids := []string{}
	for _, srcFile := range srcFiles {
		srcFilePath := filepath.Join(confCar.InputDir, srcFile.Name())
		srcFileCid, err := ipfs.IpfsUploadFileByWebApi(utils.UrlJoin(ipfsUrlPrefix, "api/v0/add?stream-channels=true&pin=true"), srcFilePath)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}

		srcFileCids = append(srcFileCids, *srcFileCid)
	}
	carFileDataCid, err := ipfs.MergeFiles2CarFile(ipfsUrlPrefix, srcFileCids)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	logs.GetLogger().Info("data CID:", *carFileDataCid)
	carFileName := *carFileDataCid + ".car"
	carFilePath := filepath.Join(confCar.OutputDir, carFileName)
	err = ipfs.Export2CarFile(ipfsUrlPrefix, *carFileDataCid, carFilePath)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	carFiles := []*libmodel.FileDesc{}
	carFile := libmodel.FileDesc{}
	carFile.CarFileName = carFileName
	carFile.CarFilePath = carFilePath

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

	carFile.DataCid = *dataCid

	carFile.CarFileSize = utils.GetFileSize(carFile.CarFilePath)

	if confCar.GenerateMd5 {
		carFileMd5, err := checksum.MD5sum(carFile.CarFilePath)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}
		carFile.CarFileMd5 = carFileMd5
	}

	carFiles = append(carFiles, &carFile)

	_, err = WriteCarFilesToFiles(carFiles, confCar.OutputDir, constants.JSON_FILE_NAME_BY_CAR, constants.CSV_FILE_NAME_BY_CAR, SUBCOMMAND_CAR)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	logs.GetLogger().Info(len(carFiles), " car files have been created to directory:", confCar.OutputDir)
	logs.GetLogger().Info("Please upload car files to web server or ipfs server.")

	return carFiles, nil
}

func getCarFile() {
	fileHash, err := ipfs.IpfsUploadFileByWebApi("http://192.168.88.41:5001/api/v0/add?stream-channels=true&pin=true", "/home/peware/swan_dora/srcFiles/gnomad.genomes.v3.1.1.sites.chr22.vcf.bgz_02.car")
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}

	logs.GetLogger().Info("source file hash:", *fileHash)

	cids := []string{
		*fileHash,
	}
	dataCid, err := ipfs.MergeFiles2CarFile("http://192.168.88.41:5001", cids)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}

	logs.GetLogger().Info("data CID:", *dataCid)
	err = ipfs.Export2CarFile("http://192.168.88.41:5001", *dataCid, "/home/peware/swan_dora/srcFiles/"+*dataCid+".car")
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}
}
