package command

import (
	"bufio"
	"fmt"
	metacar "github.com/FogMeta/meta-lib/module/ipfs"
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
	"golang.org/x/xerrors"
	"io"
	"os"
	"path/filepath"
)

var MetaCarCmd = &cli.Command{
	Name:            "meta-car",
	Usage:           "Utility tools for CAR file(s)",
	Subcommands:     []*cli.Command{metaCarCmd, getRootCmd, listCarCmd, cmdRestoreCar},
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
			Usage:    "directory where source file(s) is(are) in.",
		},
		&cli.StringFlag{
			Name:     "output-dir",
			Required: true,
			Usage:    "directory where CAR file(s) will be generated. (default: \"/tmp/tasks\")",
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
			Value: true,
		},
	},
}

var cmdRestoreCar = &cli.Command{
	Name:      "restore",
	Usage:     "Restore original files from CAR(s)",
	ArgsUsage: "[inputPath]",
	Action:    metaCarRestore,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "input-dir",
			Required: true,
			Usage:    "directory where source file(s) is(are) in.",
		},
		&cli.StringFlag{
			Name:     "output-dir",
			Required: true,
			Usage:    "directory where CAR file(s) will be generated. (default: \"/tmp/tasks\")",
		},
		&cli.IntFlag{
			Name:  "parallel",
			Value: 2,
			Usage: "specify how many number of goroutines runs when generate file node",
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
		err := fmt.Errorf("gocar file size limit is too smal")
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
	inputDir := c.String("input-dir")
	outputDir := c.String("output-dir")

	err := metacar.RestoreCar(outputDir, inputDir)
	if err != nil {
		return err
	}

	fmt.Println("Restore CAR To:", outputDir)

	return nil
}

func createFilesDesc(cmdGoCar *CmdGoCar, carInfos []metacar.CarInfo) ([]*libmodel.FileDesc, error) {

	var lotusClient *lotus.LotusClient
	var err error
	if cmdGoCar.ImportFlag {
		lotusClient, err = lotus.LotusGetClient(cmdGoCar.LotusClientApiUrl, cmdGoCar.LotusClientAccessToken)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}
	}

	fileDescs := []*libmodel.FileDesc{}
	for _, carInfo := range carInfos {

		fileDesc := libmodel.FileDesc{}
		fileDesc.PayloadCid = carInfo.RootCid
		fileDesc.CarFileName = carInfo.CarFileName
		fileDesc.CarFileUrl = carInfo.CarFileName
		fileDesc.CarFilePath = carInfo.CarFilePath

		pieceCid, pieceSize, err := calcCommP(carInfo.CarFilePath)
		if err == nil {
			carInfo.PieceCID = pieceCid
			carInfo.PieceSize = int64(pieceSize)

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

		fileDescs = append(fileDescs, &fileDesc)
	}

	_, err = WriteCarFilesToFiles(fileDescs, cmdGoCar.OutputDir, JSON_FILE_NAME_CAR_UPLOAD, CSV_FILE_NAME_CAR_UPLOAD)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

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
		return "", 0, xerrors.Errorf("not a car file: %w", err)
	}

	if _, err := rdr.Seek(0, io.SeekStart); err != nil {
		return "", 0, xerrors.Errorf("seek to start: %w", err)
	}

	pieceReader, pieceSize := padreader.New(rdr, uint64(stat.Size()))
	pieceCid, err := ffiwrapper.GeneratePieceCIDFromFile(arbitraryProofType, pieceReader, pieceSize)

	if err != nil {
		return "", 0, xerrors.Errorf("computing commP failed: %w", err)
	}

	return pieceCid.String(), uint64(pieceSize), nil
}
