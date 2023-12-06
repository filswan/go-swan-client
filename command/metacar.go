package command

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	metacar "github.com/FogMeta/meta-lib/module/ipfs"
	"github.com/FogMeta/meta-lib/util"
	"github.com/codingsince1985/checksum"
	"github.com/filecoin-project/go-commp-utils/ffiwrapper"
	"github.com/filecoin-project/go-padreader"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filswan/go-swan-lib/client/lotus"
	"github.com/filswan/go-swan-lib/logs"
	libmodel "github.com/filswan/go-swan-lib/model"
	"github.com/filswan/go-swan-lib/utils"
	"github.com/ipld/go-car"
	"github.com/urfave/cli/v2"
)

var MetaCarCmd = &cli.Command{
	Name:            "meta-car",
	Usage:           "Utility tools for CAR file(s)",
	Subcommands:     []*cli.Command{metaCarCmd, getRootCmd, listCarCmd, cmdRestoreCar, cmdExtractFile},
	HideHelpCommand: true,
}

var getRootCmd = &cli.Command{
	Name:      "root",
	Usage:     "Get a CAR's root CID",
	ArgsUsage: "filename",
	Action:    metaCarRoot,
}

var listCarCmd = &cli.Command{
	Name:      "list",
	Usage:     "List the CIDs in a CAR",
	ArgsUsage: "filename",
	Action:    metaCarList,
}

var metaCarCmd = &cli.Command{
	Name:   "generate-car",
	Usage:  "Generate CAR files of the specified size",
	Action: metaCarBuildFromDir,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "input-dir",
			Required: true,
			Usage:    "directory where source files are in.",
		},
		&cli.StringFlag{
			Name:     "output-dir",
			Required: true,
			Usage:    "directory where CAR file(s) will be generated.",
		},
		&cli.Uint64Flag{
			Name:  "slice-size",
			Value: 17179869184, // 16G
			Usage: "specify chunk piece size",
		},
		&cli.IntFlag{
			Name:  "parallel",
			Usage: "number goroutines run when building ipld nodes",
			Value: 2,
		},
		&cli.BoolFlag{
			Name:  "import",
			Usage: "whether to import CAR file to lotus",
			Value: false,
		},
	},
}

var cmdRestoreCar = &cli.Command{
	Name:      "restore",
	Usage:     "Restore original files from the CAR file",
	ArgsUsage: "[inputPath]",
	Action:    metaCarRestore,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "input-dir",
			Required: true,
			Usage:    "absolute directory to the CAR file.",
		},
		&cli.StringFlag{
			Name:     "output-dir",
			Required: true,
			Usage:    "directory where file(s) will be generated.",
		},
		&cli.IntFlag{
			Name:  "parallel",
			Value: 2,
			Usage: "specify how many number of goroutines runs when generate file node",
		},
	},
}

var cmdExtractFile = &cli.Command{
	Name:      "extract",
	Usage:     "Extract one original file from the CAR file",
	ArgsUsage: "[inputPath]",
	Action:    metaCarExtract,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "input-dir",
			Required: true,
			Usage:    "absolute directory to the CAR file.",
		},
		&cli.StringFlag{
			Name:     "file-name",
			Required: true,
			Usage:    "file name which in the CAR file.",
		},
		&cli.StringFlag{
			Name:     "output-dir",
			Required: true,
			Usage:    "directory where file will be generated.",
		},
	},
}

func metaCarList(c *cli.Context) error {
	carFile := c.Args().First()

	info, err := metacar.ListCarFile(carFile)
	if err != nil {
		return err
	}

	fmt.Println("List CAR :", carFile)
	for index, val := range info {
		fmt.Println(index, val)
	}

	return nil
}

func metaCarRoot(c *cli.Context) error {
	carFile := c.Args().First()

	root, err := metacar.GetCarRoot(carFile)
	if err != nil {
		return err
	}

	fmt.Println("CAR :", carFile)
	fmt.Println("CID :", root)
	return nil
}

func getFilesSize(args []string) (int64, error) {
	totalSize := int64(0)
	fileList, err := util.GetFileList(args)
	if err != nil {
		return int64(0), err
	}

	for _, path := range fileList {
		finfo, err := os.Stat(path)
		if err != nil {
			return int64(0), err
		}
		totalSize += finfo.Size()
	}

	return totalSize, nil
}

func metaCarBuildFromDir(c *cli.Context) error {
	outputDir := c.String("output-dir")
	sliceSize := c.Uint64("slice-size")
	inputDir := c.String("input-dir")

	cmdGoCar := GetCmdGoCar(inputDir, &outputDir, c.Int("parallel"), c.Int64("slice-size"), false, c.Bool("import"))

	err := utils.CheckDirExists(cmdGoCar.InputDir, DIR_NAME_INPUT)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	err = utils.CreateDirIfNotExists(cmdGoCar.OutputDir, DIR_NAME_OUTPUT)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	if sliceSize <= 0 {
		err := fmt.Errorf("slice size should bigger than 0")
		logs.GetLogger().Error(err)
		return err
	}

	dirSize, err := getFilesSize([]string{inputDir})
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	if uint64(dirSize) > sliceSize {
		err := fmt.Errorf("the total size of the input directory must be smaller than the slice size")
		logs.GetLogger().Error(err)
		return err
	}

	//TODO:Generate description file
	carInfos, err := metacar.GenerateCarFromDirEx(cmdGoCar.OutputDir, cmdGoCar.InputDir, cmdGoCar.GocarFileSizeLimit, true)
	if err != nil {
		return err
	}

	fileDescs, err := createFilesDesc(cmdGoCar, carInfos)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	logs.GetLogger().Info(len(fileDescs), " car files have been created to directory:", outputDir)
	logs.GetLogger().Info("Please upload car files to web server or ipfs server.")

	return nil
}

func metaCarRestore(c *cli.Context) error {
	inputDir := c.String("input-path")
	outputDir := c.String("output-dir")

	err := metacar.RestoreCar(outputDir, inputDir)
	if err != nil {
		return err
	}

	fmt.Println("Restore CAR To:", outputDir)

	return nil
}

func metaCarExtract(c *cli.Context) error {
	inputCar := c.String("input-path")

	inFileName := c.String("file-name")
	outputDir := c.String("output-dir")

	err := metacar.ExtractFileFromCar(outputDir, inputCar, inFileName)
	if err != nil {
		return err
	}

	fmt.Println("Extract ", inFileName, " To:", outputDir)

	return nil
}

func createFilesDesc(cmdGoCar *CmdGoCar, carInfos []metacar.CarInfo) ([]*MetaFileDesc, error) {

	var lotusClient *lotus.LotusClient
	var err error
	if cmdGoCar.ImportFlag {
		lotusClient, err = lotus.LotusGetClient(cmdGoCar.LotusClientApiUrl, cmdGoCar.LotusClientAccessToken)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}
	}

	fileDescs := []*MetaFileDesc{}
	for i, carInfo := range carInfos {

		fileDesc := MetaFileDesc{}
		fileDesc.PayloadCid = carInfo.RootCid
		fileDesc.CarFileName = carInfo.CarFileName
		fileDesc.CarFileUrl = carInfo.CarFileName
		fileDesc.CarFilePath = carInfo.CarFilePath

		pieceCid, pieceSize, err := calcCommP(carInfo.CarFilePath)
		if err == nil {
			carInfo.PieceCID = pieceCid
			carInfo.PieceSize = int64(pieceSize)

			carInfos[i].PieceCID = pieceCid
			carInfos[i].PieceSize = int64(pieceSize)

		}
		fileDesc.PieceCid = carInfo.PieceCID
		fileDesc.CarFileSize = carInfo.PieceSize

		if cmdGoCar.ImportFlag {
			dataCid, err := lotusClient.LotusClientImport(fileDesc.CarFilePath, true)
			if err != nil {
				logs.GetLogger().Error(err)
				return nil, err
			}

			fileDesc.PayloadCid = *dataCid
		}

		//TODO: source files
		fileDesc.SourceFileName = filepath.Base(cmdGoCar.InputDir)
		fileDesc.SourceFilePath = cmdGoCar.InputDir
		for _, detail := range carInfo.Details {
			fileDesc.SourceFileSize = fileDesc.SourceFileSize + detail.FileSize
		}

		if cmdGoCar.GenerateMd5 {
			//TODO: source filess
			if utils.IsFileExistsFullPath(fileDesc.SourceFilePath) {
				srcFileMd5, err := checksum.MD5sum(fileDesc.SourceFilePath)
				if err != nil {
					logs.GetLogger().Error(err)
					return nil, err
				}
				fileDesc.SourceFileMd5 = srcFileMd5
			}

			carFileMd5, err := checksum.MD5sum(fileDesc.CarFilePath)
			if err != nil {
				logs.GetLogger().Error(err)
				return nil, err
			}
			fileDesc.CarFileMd5 = carFileMd5
		}

		fileDesc.SrcDetail = append(fileDesc.SrcDetail, carInfo.Details...)
		// logs.GetLogger().Info("Details:", carInfo.Details)
		// fileDesc.SrcDetail = carInfo.Details

		fileDescs = append(fileDescs, &fileDesc)
	}

	_, err = metaWriteCarFilesToFiles(fileDescs, cmdGoCar.OutputDir, JSON_FILE_NAME_CAR_UPLOAD, CSV_FILE_NAME_CAR_UPLOAD)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	//_, err = metaWriteIndexToJsonFile(carInfos, cmdGoCar.OutputDir, INDEX_FILE_NAME_CAR_UPLOAD)
	//if err != nil {
	//	logs.GetLogger().Error(err)
	//	return nil, err
	//}

	return fileDescs, nil
}

func calcCommP(inputCarFile string) (string, uint64, error) {

	arbitraryProofType := abi.RegisteredSealProof_StackedDrg32GiBV1_1

	rdr, err := os.Open(inputCarFile)
	if err != nil {
		return "", 0, err
	}
	defer rdr.Close() //nolint:errcheck

	stat, err := rdr.Stat()
	if err != nil {
		return "", 0, err
	}

	// check that the data is a car file; if it's not, retrieval won't work
	_, err = car.ReadHeader(bufio.NewReader(rdr))
	if err != nil {
		return "", 0, fmt.Errorf("not a car file: %w", err)
	}

	if _, err := rdr.Seek(0, io.SeekStart); err != nil {
		return "", 0, fmt.Errorf("seek to start: %w", err)
	}

	pieceReader, pieceSize := padreader.New(rdr, uint64(stat.Size()))
	pieceCid, err := ffiwrapper.GeneratePieceCIDFromFile(arbitraryProofType, pieceReader, pieceSize)

	if err != nil {
		return "", 0, fmt.Errorf("computing commP failed: %w", err)
	}

	return pieceCid.String(), uint64(pieceSize), nil
}

type MetaFileDesc struct {
	Uuid           string
	SourceFileName string
	SourceFilePath string
	SourceFileMd5  string
	SourceFileSize int64
	CarFileName    string
	CarFilePath    string
	CarFileMd5     string
	CarFileUrl     string
	CarFileSize    int64
	PayloadCid     string
	PieceCid       string
	StartEpoch     *int64
	SourceId       *int
	Deals          []*libmodel.DealInfo
	SrcDetail      []metacar.DetailInfo `json:"src_detail"`
}

func metaWriteCarFilesToFiles(carFiles []*MetaFileDesc, outputDir, jsonFilename, csvFileName string) (*string, error) {
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	jsonFilePath, err := metaWriteFileDescsToJsonFile(carFiles, outputDir, jsonFilename)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	err = metaWriteCarFilesToCsvFile(carFiles, outputDir, csvFileName)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	return jsonFilePath, nil
}

func metaWriteIndexToJsonFile(carInfos []metacar.CarInfo, outputDir, indexFileName string) (*string, error) {
	jsonFilePath := filepath.Join(outputDir, indexFileName)
	content, err := json.MarshalIndent(carInfos, "", " ")
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	err = os.WriteFile(jsonFilePath, content, 0644)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	logs.GetLogger().Info("Metadata index file generated: ", jsonFilePath)
	return &jsonFilePath, nil
}

func metaWriteFileDescsToJsonFile(fileDescs []*MetaFileDesc, outputDir, jsonFileName string) (*string, error) {
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

func metaWriteCarFilesToCsvFile(carFiles []*MetaFileDesc, outDir, csvFileName string) error {
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
	headers = append(headers, "src_detail")

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

		if len(carFile.SrcDetail) > 0 {
			detailByte, err := json.Marshal(carFile.SrcDetail)
			if err != nil {
				logs.GetLogger().Error(err)
				columns = append(columns, "")
			} else {
				columns = append(columns, string(detailByte))
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
