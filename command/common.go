package command

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/filedrive-team/go-graphsplit"
	"github.com/filswan/go-swan-lib/logs"
	libmodel "github.com/filswan/go-swan-lib/model"
	"github.com/google/uuid"
	"github.com/ipld/go-car"
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
	CMD_VERSION    = "version"

	JSON_FILE_NAME_CAR_UPLOAD = "car.json"
	JSON_FILE_NAME_TASK       = "-metadata.json"
	JSON_FILE_NAME_DEAL       = "-deals.json"
	JSON_FILE_NAME_DEAL_AUTO  = "-auto-deals.json"

	CSV_FILE_NAME_CAR_UPLOAD = "car.csv"
	CSV_FILE_NAME_TASK       = "-metadata.csv"
	CSV_FILE_NAME_DEAL       = "-deals.csv"
	CSV_FILE_NAME_DEAL_AUTO  = "-auto-deals.csv"

	INDEX_FILE_NAME_CAR_UPLOAD = "car.idx"

	DIR_NAME_INPUT  = "input"
	DIR_NAME_OUTPUT = "output"

	DURATION_MIN = 518400
	DURATION_MAX = 1555200
	VERSION      = "v2.3.0"
)

var publicChain = map[string][]string{
	"1":  {"https://eth-mainnet.gateway.pokt.network/v1/5f3453978e354ab992c4da79", "https://eth-rpc.gateway.pokt.network"},                      // 1     Ethereum Mainnet
	"2":  {"https://bsc-mainnet.gateway.pokt.network/v1/lb/6136201a7bad1500343e248d"},                                                           // 56    Binance Smart Chain Mainnet
	"3":  {"https://api.avax.network/ext/bc/C/rpc", "https://avax-mainnet.gateway.pokt.network/v1/lb/605238bf6b986eea7cf36d5e/ext/bc/C/rpc"},    // 43114 Avalanche C-Chain
	"4":  {"https://poly-rpc.gateway.pokt.network"},                                                                                             // 137   Polygon Mainnet
	"5":  {"https://fantom-mainnet.gateway.pokt.network/v1/lb/6261a8a154c745003bcdb0f8"},                                                        // 250   Fantom Opera
	"6":  {"https://xdai-rpc.gateway.pokt.network"},                                                                                             // 100   Gnosis Chain (formerly xDai)
	"7":  {"https://pokt-api.iotex.io", "https://iotex-mainnet.gateway.pokt.network/v1/lb/6176f902e19001003499f492"},                            // 4689  IoTeX Network Mainnet
	"8":  {"https://harmony-0-rpc.gateway.pokt.network"},                                                                                        // 1666600000 Harmony Mainnet Shard 0
	"9":  {"https://boba-mainnet.gateway.pokt.network/v1/lb/623ad21b20354900396fed7f"},                                                          // 288   Boba Network
	"10": {"https://fuse-rpc.gateway.pokt.network"},                                                                                             // 122   Fuse Mainnet
	"11": {"https://avax-dfk.gateway.pokt.network/v1/lb/6244818c00b9f0003ad1b619/ext/bc/q2aTwKuyzgs8pynF7UXBZCU7DejbZbZ6EUyHr3JQzYgwNPUPi/rpc"}, // 53935  DFK Chain
	"12": {"https://evmos-mainnet.gateway.pokt.network/v1/lb/627586ddea1b320039c95205"},                                                         // 9001   Evmos
	"13": {"https://avax-cra-rpc.gateway.pokt.network"},                                                                                         // 73772  Swimmer Network
}

func WriteCarFilesToFiles(carFiles []*libmodel.FileDesc, outputDir, jsonFilename, csvFileName string) (*string, error) {
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	jsonFilePath, err := WriteFileDescsToJsonFile(carFiles, outputDir, jsonFilename)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	err = WriteCarFilesToCsvFile(carFiles, outputDir, csvFileName)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return jsonFilePath, nil
}

func WriteFileDescsToJsonFile(fileDescs []*libmodel.FileDesc, outputDir, jsonFileName string) (*string, error) {
	jsonFilePath := filepath.Join(outputDir, jsonFileName)
	content, err := json.MarshalIndent(fileDescs, "", " ")
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	err = os.WriteFile(jsonFilePath, content, 0644)
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
	contents, err := os.ReadFile(jsonFilePath)
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

func WriteCarFilesToCsvFile(carFiles []*libmodel.FileDesc, outDir, csvFileName string) error {
	csvFilePath := filepath.Join(outDir, csvFileName)
	var headers []string
	headers = append(headers, "uuid")
	headers = append(headers, "source_file_name")
	headers = append(headers, "source_file_path")
	headers = append(headers, "source_file_md5")
	headers = append(headers, "source_file_size")
	headers = append(headers, "car_file_name")
	headers = append(headers, "car_file_path")
	headers = append(headers, "car_file_md5")
	headers = append(headers, "car_file_url")
	headers = append(headers, "car_file_size")
	headers = append(headers, "pay_load_cid")
	headers = append(headers, "piece_cid")
	headers = append(headers, "start_epoch")
	headers = append(headers, "source_id")
	headers = append(headers, "deals")

	file, err := os.Create(csvFilePath)
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
		columns = append(columns, carFile.Uuid)
		columns = append(columns, carFile.SourceFileName)
		columns = append(columns, carFile.SourceFilePath)
		columns = append(columns, carFile.SourceFileMd5)
		columns = append(columns, strconv.FormatInt(carFile.SourceFileSize, 10))
		columns = append(columns, carFile.CarFileName)
		columns = append(columns, carFile.CarFilePath)
		columns = append(columns, carFile.CarFileMd5)
		columns = append(columns, carFile.CarFileUrl)
		columns = append(columns, strconv.FormatInt(carFile.CarFileSize, 10))
		columns = append(columns, carFile.PayloadCid)
		columns = append(columns, carFile.PieceCid)

		if carFile.StartEpoch != nil {
			columns = append(columns, strconv.FormatInt(*carFile.StartEpoch, 10))
		} else {
			columns = append(columns, "")
		}

		if carFile.SourceId != nil {
			columns = append(columns, strconv.Itoa(*carFile.SourceId))
		} else {
			columns = append(columns, "")
		}
		if len(carFile.Deals) > 0 {
			dealsByte, err := json.Marshal(carFile.Deals)
			if err != nil {
				logs.GetLogger().Error(err)
				columns = append(columns, "")
			} else {
				columns = append(columns, string(dealsByte))
			}
		} else {
			columns = append(columns, "")
		}
		err = writer.Write(columns)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
	}

	logs.GetLogger().Info("Metadata csv generated: ", csvFilePath)

	return nil
}

func ReadFileFromCsvFile(inputDir, csvFilename string) ([]*libmodel.FileDesc, error) {
	csvFilePath := filepath.Join(inputDir, csvFilename)
	fileDescs, err := ReadFileFromCsvFileByFullPath(csvFilePath)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return fileDescs, nil
}

func ReadFileFromCsvFileByFullPath(csvFileName string) (fileDesc []*libmodel.FileDesc, err error) {
	fs, err := os.Open(csvFileName)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}
	defer fs.Close()
	r := csv.NewReader(fs)
	isFirst := true
	for {
		row, err := r.Read()
		if err != nil && err != io.EOF {
			logs.GetLogger().Error(err)
			break
		}
		if err == io.EOF {
			break
		}
		if isFirst {
			isFirst = false
			continue
		}

		carFile := new(libmodel.FileDesc)
		if len(row) == 5 {
			carFile.Uuid = uuid.NewString()
			carFile.SourceFileName = row[0]
			carFile.SourceFilePath = row[4]
			carFile.SourceFileMd5 = ""
			SourceFileSize, _ := strconv.ParseInt(row[1], 10, 64)
			carFile.SourceFileSize = SourceFileSize
			carFile.CarFileName = row[0]
			carFile.CarFilePath = row[4]
			carFile.CarFileMd5 = ""
			carFile.CarFileUrl = row[4]
			carFile.CarFileSize = SourceFileSize
			carFile.PayloadCid = row[3]
			carFile.PieceCid = row[2]
		}

		if len(row) == 11 {
			carFile.Uuid = uuid.NewString()
			carFile.SourceFileName = row[6]
			carFile.SourceFilePath = row[7]
			carFile.SourceFileMd5 = row[9]
			SourceFileSize, _ := strconv.ParseInt(row[8], 10, 64)
			carFile.SourceFileSize = SourceFileSize
			carFile.CarFileName = row[0]
			carFile.CarFilePath = row[1]
			carFile.CarFileMd5 = row[5]
			carFile.CarFileUrl = row[10]
			CarFileSize, _ := strconv.ParseInt(row[4], 10, 64)
			carFile.CarFileSize = CarFileSize
			carFile.PayloadCid = row[3]
			carFile.PieceCid = row[2]
		}

		if len(row) > 11 {
			var sourceFileSize, carFileSize, startEpochs int64
			var sourceIds int
			if len(row[4]) > 0 {
				sourceFileSize, _ = strconv.ParseInt(row[4], 10, 64)
			}
			if len(row[9]) > 0 {
				carFileSize, _ = strconv.ParseInt(row[9], 10, 64)
			}
			if len(row[12]) > 0 {
				startEpochs, _ = strconv.ParseInt(row[12], 10, 64)
			}
			if len(row[13]) > 0 {
				sourceIds, _ = strconv.Atoi(row[13])
			}
			dealInfo := []*libmodel.DealInfo{}
			if len(row[14]) > 0 {
				err = json.Unmarshal([]byte(row[14]), &dealInfo)
				if err != nil {
					println(err)
				}
			}

			carFile.Uuid = row[0]
			carFile.SourceFileName = row[1]
			carFile.SourceFilePath = row[2]
			carFile.SourceFileMd5 = row[3]
			carFile.SourceFileSize = sourceFileSize
			carFile.CarFileName = row[5]
			carFile.CarFilePath = row[6]
			carFile.CarFileMd5 = row[7]
			carFile.CarFileUrl = row[8]
			carFile.CarFileSize = carFileSize
			carFile.PayloadCid = row[10]
			carFile.PieceCid = row[11]
			carFile.StartEpoch = &startEpochs
			carFile.SourceId = &sourceIds
			carFile.Deals = dealInfo
		}
		if carFile != nil {
			fileDesc = append(fileDesc, carFile)
		}
	}
	return
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

func CalculateValueByCarFile(carFilePath string, playload, piece bool) (string, string, uint64, error) {
	var dataCid string
	if playload {
		f, err := os.Open(carFilePath)
		if err != nil {
			return "", "", 0, fmt.Errorf("failed to open CAR file: %w", err)
		}
		defer f.Close() //nolint:errcheck

		hd, err := car.ReadHeader(bufio.NewReader(f))
		if err != nil {
			return "", "", 0, fmt.Errorf("failed to read CAR header: %w", err)
		}
		if len(hd.Roots) != 1 {
			return "", "", 0, errors.New("car file can have one and only one header")
		}
		if hd.Version != 1 && hd.Version != 2 {
			return "", "", 0, fmt.Errorf("car version must be 1 or 2, is %d", hd.Version)
		}
		dataCid = hd.Roots[0].String()
	}

	var pieceCid string
	var pieceSize uint64
	if piece {
		cpRes, err := graphsplit.CalcCommP(context.TODO(), carFilePath)
		if err != nil {
			logs.GetLogger().Error(err)
		}
		pieceCid = cpRes.Root.String()
		pieceSize = uint64(cpRes.Size)
	}
	return dataCid, pieceCid, pieceSize, nil

}
