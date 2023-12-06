package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	big2 "github.com/filecoin-project/go-state-types/big"
	"github.com/filswan/go-swan-client/command"
	"github.com/filswan/go-swan-client/config"
	"github.com/filswan/go-swan-lib/client/lotus"
	"github.com/filswan/go-swan-lib/client/swan"
	"github.com/filswan/go-swan-lib/constants"
	libconstants "github.com/filswan/go-swan-lib/constants"
	"github.com/filswan/go-swan-lib/logs"
	"github.com/filswan/go-swan-lib/utils"
	boost "github.com/filswan/swan-boost-lib/client"
	"github.com/julienschmidt/httprouter"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:                 "swan-client",
		Usage:                "A PiB level data onboarding tool for Filecoin Network",
		Version:              command.VERSION,
		EnableBashCompletion: true,
		After: func(context *cli.Context) error {
			if r := recover(); r != nil {
				panic(r)
			}
			return nil
		},
		Commands:        []*cli.Command{daemonCmd, walletCmd, toolsCmd, command.MetaCarCmd, uploadCmd, taskCmd, dealCmd, autoCmd, calculateCmd, rpcApiCmd, rpcCmd},
		HideHelpCommand: true,
	}
	if err := app.Run(os.Args); err != nil {
		var phe *PrintHelpErr
		fmt.Fprintf(os.Stderr, "ERROR: %s\n\n", err)
		if errors.As(err, &phe) {
			_ = cli.ShowCommandHelp(phe.Ctx, phe.Ctx.Command.Name)
		}
		os.Exit(1)
	}
}

var walletCmd = &cli.Command{
	Name:  "wallet",
	Usage: "Manage wallets with Swan client",
	Subcommands: []*cli.Command{
		walletNewCmd,
		walletListCmd,
		walletImportCmd,
		walletExportCmd,
		walletDeleteCmd,
	},
	HideHelpCommand: true,
}

var walletNewCmd = &cli.Command{
	Name:      "new",
	Usage:     "Generate a new key of the given type",
	ArgsUsage: "[bls|secp256k1 (default secp256k1)]",
	Action: func(cctx *cli.Context) error {
		t := cctx.Args().First()
		if t == "" {
			t = libconstants.WALLET_TYPE_256
		}
		marketVersion := strings.TrimSpace(config.GetConfig().Main.MarketVersion)
		if marketVersion == constants.MARKET_VERSION_2 {
			if err := checkRepo(); err != nil {
				return err
			}
			return boost.GetClient(config.GetConfig().Main.SwanRepo).WalletNew(t)
		} else {
			return errors.New("only support market_version=“1.2” ")
		}
	},
}

var walletListCmd = &cli.Command{
	Name:  "list",
	Usage: "List wallet address",
	Action: func(ctx *cli.Context) error {
		marketVersion := strings.TrimSpace(config.GetConfig().Main.MarketVersion)
		if marketVersion == constants.MARKET_VERSION_2 {
			if err := checkRepo(); err != nil {
				return err
			}
			lotusClient, err := lotus.LotusGetClient(config.GetConfig().Lotus.ClientApiUrl, config.GetConfig().Lotus.ClientAccessToken)
			if err != nil {
				logs.GetLogger().Error(err)
				return err
			}
			return boost.GetClient(config.GetConfig().Main.SwanRepo).WithClient(lotusClient).WalletList()

		} else {
			return errors.New("only support market_version=“1.2” ")
		}
	},
}

var walletImportCmd = &cli.Command{
	Name:      "import",
	Usage:     "import keys",
	ArgsUsage: "[<path> (optional, will read from stdin if omitted)]",
	Action: func(ctx *cli.Context) error {
		marketVersion := strings.TrimSpace(config.GetConfig().Main.MarketVersion)
		if marketVersion == constants.MARKET_VERSION_2 {
			if err := checkRepo(); err != nil {
				return err
			}
			var inputData []byte
			if !ctx.Args().Present() || ctx.Args().First() == "-" {
				reader := bufio.NewReader(os.Stdin)
				fmt.Print("Enter private key: ")
				indata, err := reader.ReadBytes('\n')
				if err != nil {
					return err
				}
				inputData = indata
			} else {
				fdata, err := os.ReadFile(ctx.Args().First())
				if err != nil {
					return err
				}
				inputData = fdata
			}
			lotusClient, err := lotus.LotusGetClient(config.GetConfig().Lotus.ClientApiUrl, config.GetConfig().Lotus.ClientAccessToken)
			if err != nil {
				logs.GetLogger().Error(err)
				return err
			}
			return boost.GetClient(config.GetConfig().Main.SwanRepo).WithClient(lotusClient).WalletImport(inputData)

		} else {
			return errors.New("only support market_version=“1.2” ")
		}
	},
}

var walletExportCmd = &cli.Command{
	Name:      "export",
	Usage:     "export keys",
	ArgsUsage: "[address]",
	Action: func(cctx *cli.Context) error {
		if !cctx.Args().Present() || cctx.NArg() != 1 {
			return fmt.Errorf("must specify address to export")
		}
		addr := cctx.Args().First()

		marketVersion := strings.TrimSpace(config.GetConfig().Main.MarketVersion)
		if marketVersion == constants.MARKET_VERSION_2 {
			if err := checkRepo(); err != nil {
				return err
			}
			return boost.GetClient(config.GetConfig().Main.SwanRepo).WalletExport(addr)
		} else {
			return errors.New("only support market_version=“1.2” ")
		}
	},
}

var walletDeleteCmd = &cli.Command{
	Name:      "delete",
	Usage:     "Delete an account from the wallet",
	ArgsUsage: "<address> ",
	Action: func(cctx *cli.Context) error {
		if !cctx.Args().Present() || cctx.NArg() != 1 {
			return fmt.Errorf("must specify address to delete")
		}
		addr := cctx.Args().First()

		marketVersion := strings.TrimSpace(config.GetConfig().Main.MarketVersion)
		if marketVersion == constants.MARKET_VERSION_2 {
			if err := checkRepo(); err != nil {
				return err
			}
			return boost.GetClient(config.GetConfig().Main.SwanRepo).WalletDelete(addr)
		} else {
			return errors.New("only support market_version=“1.2” ")
		}
	},
}

var daemonCmd = &cli.Command{
	Name:  "daemon",
	Usage: "Start a API service process",
	Before: func(context *cli.Context) error {
		swanClient, err := swan.GetClient(config.GetConfig().Main.SwanApiUrl, config.GetConfig().Main.SwanApiKey,
			config.GetConfig().Main.SwanAccessToken, "")
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		swanClient.StatisticsNodeStatus()

		return nil
	},
	Action: func(ctx *cli.Context) error {
		router := httprouter.New()
		router.POST("/chain/rpc", func(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
			defer request.Body.Close()
			data, _ := io.ReadAll(request.Body)
			var reqParam struct {
				ChainId string `json:"chain_id"`
				Params  string `json:"params"`
			}
			if err := json.Unmarshal(data, &reqParam); err != nil {
				logs.GetLogger().Error(err)
				writer.WriteHeader(http.StatusBadRequest)
				writer.Write([]byte("request parameter is bad"))
				return
			}
			if reqParam.ChainId == "" && reqParam.Params == "" {
				logs.GetLogger().Error("both chain-id and params are required")
				writer.WriteHeader(http.StatusNotFound)
				writer.Write([]byte("both chain-id and params are required"))
				return
			}
			result, err := command.SendRpcReqAndResp(reqParam.ChainId, reqParam.Params)
			if err != nil {
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte(err.Error()))
				return
			}
			writer.Write(result)
			go func() {
				swanClient, err := swan.GetClient(config.GetConfig().Main.SwanApiUrl, config.GetConfig().Main.SwanApiKey,
					config.GetConfig().Main.SwanAccessToken, "")
				if err != nil {
					logs.GetLogger().Error(err)
					return
				}
				swanClient.StatisticsChainInfo(reqParam.ChainId)
			}()
		})

		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					panic(r)
				}
			}()
			logs.GetLogger().Infof("listen port: 8099")
			http.ListenAndServe(":8099", router)
		}()

		select {
		case sig := <-c:
			logs.GetLogger().Warnf(" receive %s signal exit.", sig)
		}
		return nil
	},
}

var uploadCmd = &cli.Command{
	Name:      "upload",
	Usage:     "Upload CAR file to ipfs server",
	ArgsUsage: "[inputPath]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "input-dir",
			Aliases: []string{"i"},
			Usage:   "directory where source files are in.",
		},
	},
	Action: func(ctx *cli.Context) error {
		inputDir := ctx.String("input-dir")
		if inputDir == "" {
			return errors.New("input-dir is required")
		}
		_, err := command.UploadCarFilesByConfig(inputDir)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		return nil
	},
}

var taskCmd = &cli.Command{
	Name:  "task",
	Usage: "Send task to swan",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "name",
			Usage: "task name",
		},
		&cli.StringFlag{
			Name:    "input-dir",
			Aliases: []string{"i"},
			Usage:   "absolute path where the json or csv format source files",
		},
		&cli.StringFlag{
			Name:    "out-dir",
			Aliases: []string{"o"},
			Usage:   "directory where target files will in",
			Value:   "/tmp/tasks",
		},
		&cli.BoolFlag{
			Name:  "auto-bid",
			Usage: "send the auto-bid task",
			Value: false,
		},
		&cli.BoolFlag{
			Name:  "manual-bid",
			Usage: "send the manual-bid task",
			Value: false,
		},
		&cli.StringFlag{
			Name:  "miners",
			Usage: "minerID is required when send private task (pass comma separated array of minerIDs)",
		},
		&cli.StringFlag{
			Name:  "dataset",
			Usage: "curated dataset",
		},
		&cli.StringFlag{
			Name:    "description",
			Aliases: []string{"d"},
			Usage:   "task description",
		},
		&cli.IntFlag{
			Name:    "max-copy-number",
			Aliases: []string{"max"},
			Usage:   "max copy numbers when send auto-bid or manual-bid task",
			Value:   1,
		},
	},
	Action: func(ctx *cli.Context) error {
		repoPath := config.GetConfig().Main.SwanRepo
		wallet := config.GetConfig().Sender.Wallet
		var first bool
		if _, err := os.Stat(repoPath); err != nil {
			first = true
		}
		if err := checkRepo(); err != nil {
			return err
		}
		if strings.TrimSpace(config.GetConfig().Main.MarketVersion) == constants.MARKET_VERSION_2 {
			if exist := boost.GetClient(repoPath).ValidateExistWalletAddress(wallet); !exist {
				if !first {
					return fmt.Errorf("the current client wallet address is: %s, please use the command <./swan-client wallet import wallet.key> to import the wallet private key", wallet)
				} else {
					return nil
				}
			}
		}

		inputDir := ctx.String("input-dir")
		if inputDir == "" {
			return errors.New("input-dir is required")
		}
		if !strings.HasSuffix(inputDir, "csv") && !strings.HasSuffix(inputDir, "json") {
			return errors.New("inputDir must be json or csv format file")
		}
		logs.GetLogger().Info("your input source file as: ", inputDir)

		auto := ctx.Bool("auto-bid")
		manual := ctx.Bool("manual-bid")
		minerId := ctx.String("miners")

		if auto && minerId != "" {
			return errors.New("miners need not to set when auto value is true")
		}

		if manual && minerId != "" {
			return errors.New("miners need not to set when manual value is true")
		}

		if !auto && !manual && minerId == "" {
			return errors.New("only one argument can be set among auto-bid, manual-bid and miners")
		}

		if auto && manual {
			return errors.New("auto-bid and manual-bid cannot be set at the same time")
		}

		outputDir := ctx.String("out-dir")
		_, fileDesc, _, total, err := command.CreateTaskByConfig(inputDir, &outputDir, ctx.String("name"), minerId,
			ctx.String("dataset"), ctx.String("description"), auto, manual, ctx.Int("max-copy-number"))
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}

		if auto {
			taskId := fileDesc[0].Uuid
			exitCh := make(chan interface{})
			go func() {
				defer func() {
					exitCh <- struct{}{}
				}()
				command.GetCmdAutoDeal(&outputDir).SendAutoBidDealsBySwanClientSourceId(inputDir, taskId, total)
			}()
			<-exitCh
		}
		return nil
	},
}

var dealCmd = &cli.Command{
	Name:  "deal",
	Usage: "Send manual-bid deal",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "csv",
			Usage: "the CSV file path of deal metadata",
		},
		&cli.StringFlag{
			Name:  "json",
			Usage: "the JSON file path of deal metadata",
		},
		&cli.StringFlag{
			Name:    "out-dir",
			Aliases: []string{"o"},
			Usage:   "directory where target files will in",
			Value:   "/tmp/tasks",
		},
		&cli.StringFlag{
			Name:  "miners",
			Usage: "miners must have been assigned the task (pass comma separated array of minerIDs)",
		},
	},
	Action: func(ctx *cli.Context) error {
		repoPath := config.GetConfig().Main.SwanRepo
		wallet := config.GetConfig().Sender.Wallet
		var first bool
		if _, err := os.Stat(repoPath); err != nil {
			first = true
		}
		if err := checkRepo(); err != nil {
			return err
		}
		if strings.TrimSpace(config.GetConfig().Main.MarketVersion) == constants.MARKET_VERSION_2 {
			if exist := boost.GetClient(repoPath).ValidateExistWalletAddress(wallet); !exist {
				if !first {
					return fmt.Errorf("the current client wallet address is: %s, please use the command <./swan-client wallet import wallet.key> to import the wallet private key", wallet)
				} else {
					return nil
				}
			}
		}

		metadataJsonPath := ctx.String("json")
		metadataCsvPath := ctx.String("csv")

		if len(metadataJsonPath) == 0 && len(metadataCsvPath) == 0 {
			return errors.New("at least one argument can be set between csv and json")
		}

		if len(metadataJsonPath) > 0 && len(metadataCsvPath) > 0 {
			return errors.New("metadata file path is required, it cannot contain csv file path or json file path at the same time")
		}

		if len(metadataJsonPath) > 0 {
			logs.GetLogger().Info("Metadata json file:", metadataJsonPath)
		}
		if len(metadataCsvPath) > 0 {
			logs.GetLogger().Info("Metadata csv file:", metadataCsvPath)
		}
		minerIds := ctx.String("miners")
		if _, err := command.SendDealsByConfig(ctx.String("out-dir"), minerIds, metadataJsonPath, metadataCsvPath); err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		return nil
	},
}

var autoCmd = &cli.Command{
	Name:      "auto",
	Usage:     "Auto send bid deal",
	ArgsUsage: "[inputPath]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "out-dir",
			Aliases: []string{"o"},
			Usage:   "directory where target files will in.",
		},
	},
	Action: func(ctx *cli.Context) error {
		repoPath := config.GetConfig().Main.SwanRepo
		wallet := config.GetConfig().Sender.Wallet
		var first bool
		if _, err := os.Stat(repoPath); err != nil {
			first = true
		}
		if err := checkRepo(); err != nil {
			return err
		}
		if strings.TrimSpace(config.GetConfig().Main.MarketVersion) == constants.MARKET_VERSION_2 {
			if exist := boost.GetClient(repoPath).ValidateExistWalletAddress(wallet); !exist {
				if !first {
					return fmt.Errorf("the current client wallet address is: %s, please use the command <./swan-client wallet import wallet.key> to import the wallet private key", wallet)
				} else {
					return nil
				}
			}
		}

		if err := command.SendAutoBidDealsLoopByConfig(ctx.String("out-dir")); err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		return nil
	},
	Hidden: true,
}

var inPutFlag = cli.StringFlag{
	Name:    "input-dir",
	Aliases: []string{"i"},
	Usage:   "directory where source file(s) is(are) in.",
}
var importFlag = cli.BoolFlag{
	Name:  "import",
	Usage: "whether to import CAR file to lotus",
	Value: false,
}

var outPutFlag = cli.StringFlag{
	Name:    "out-dir",
	Aliases: []string{"o"},
	Usage:   "directory where CAR file(s) will be generated.",
	Value:   "/tmp/tasks",
}

var lotusCarCmd = &cli.Command{
	Name:      "lotus",
	Usage:     "Use lotus api to generate CAR file",
	ArgsUsage: "[inputPath]",
	Flags: []cli.Flag{
		&inPutFlag,
		&outPutFlag,
		&importFlag,
	},
	Action: func(ctx *cli.Context) error {
		inputDir := ctx.String("input-dir")
		if inputDir == "" {
			return errors.New("input-dir is required")
		}
		outputDir := ctx.String("out-dir")
		if _, err := command.CreateCarFilesByConfig(inputDir, &outputDir, ctx.Bool("import")); err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		return nil
	},
}

var splitCarCmd = &cli.Command{
	Name:            "graphsplit",
	Usage:           "Use go-graphsplit tools",
	Subcommands:     []*cli.Command{generateCarCmd, carRestoreCmd},
	HideHelpCommand: true,
}

var generateCarCmd = &cli.Command{
	Name:      "car",
	Usage:     "Generate CAR files of the specified size",
	ArgsUsage: "[inputPath]",
	Flags: []cli.Flag{
		&inPutFlag,
		&outPutFlag,
		&importFlag,
		&cli.IntFlag{
			Name:  "parallel",
			Usage: "number goroutines run when building ipld nodes",
			Value: 5,
		},
		&cli.Int64Flag{
			Name:    "slice-size",
			Aliases: []string{"size"},
			Usage:   "bytes of each piece",
			Value:   17179869184,
		},
		&cli.BoolFlag{
			Name:  "parent-path",
			Usage: "generate CAR file based on whole folder",
			Value: true,
		},
	},
	Action: func(ctx *cli.Context) error {
		inputDir := ctx.String("input-dir")
		if inputDir == "" {
			return errors.New("input-dir is required")
		}
		outputDir := ctx.String("out-dir")
		if _, err := command.CreateGoCarFilesByConfig(inputDir, &outputDir, ctx.Int("parallel"), ctx.Int64("slice-size"), ctx.Bool("parent-path"), ctx.Bool("import")); err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		return nil
	},
}

var carRestoreCmd = &cli.Command{
	Name:      "restore",
	Usage:     "Restore files from CAR files",
	ArgsUsage: "[inputPath]",
	Flags: []cli.Flag{
		&outPutFlag,
		&cli.StringFlag{
			Name:     "input-dir",
			Aliases:  []string{"i"},
			Usage:    "specify source CAR path, directory or file",
			Required: true,
		},
		&cli.Int64Flag{
			Name:  "parallel",
			Usage: "number goroutines run when building ipld nodes",
			Value: 5,
		},
	},
	Action: func(ctx *cli.Context) error {
		inputDir := ctx.String("input-dir")
		if inputDir == "" {
			return errors.New("input-dir is required")
		}
		outputDir := ctx.String("out-dir")
		if err := command.RestoreCarFilesByConfig(inputDir, &outputDir, ctx.Int("parallel")); err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		return nil
	},
}

var toolsCmd = &cli.Command{
	Name:            "generate-car",
	Usage:           "Generate CAR files from a file or directory",
	Subcommands:     []*cli.Command{splitCarCmd, lotusCarCmd, ipfsCmdCarCmd},
	HideHelpCommand: true,
}

var ipfsCmdCarCmd = &cli.Command{
	Name:      "ipfs-car",
	Usage:     "use the ipfs-car command to generate the CAR file",
	ArgsUsage: "[inputPath]",
	Flags: []cli.Flag{
		&inPutFlag,
		&outPutFlag,
		&importFlag,
	},
	Action: func(ctx *cli.Context) error {
		inputDir := ctx.String("input-dir")
		if inputDir == "" {
			return errors.New("input-dir is required")
		}
		outputDir := ctx.String("out-dir")
		if _, err := command.CreateIpfsCmdCarFilesByConfig(inputDir, &outputDir, ctx.Bool("import")); err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		return nil
	},
}

var calculateCmd = &cli.Command{
	Name:      "commP",
	Usage:     "Calculate the dataCid, pieceCid, pieceSize of the CAR file",
	ArgsUsage: "[inputPath]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "car-path",
			Usage: "absolute path to the car file",
		},
		&cli.BoolFlag{
			Name:   "data-cid",
			Usage:  "whether to generate the dataCid flag",
			Value:  true,
			Hidden: true,
		},
		&cli.BoolFlag{
			Name:  "piece-cid",
			Usage: "whether to generate the pieceCid flag",
			Value: false,
		},
	},
	Action: func(ctx *cli.Context) error {
		carPath := ctx.String("car-path")
		if carPath == "" {
			return errors.New("car-path is required")
		}
		dataCid, pieceCid, pieceSize, err := command.CalculateValueByCarFile(carPath, ctx.Bool("data-cid"), ctx.Bool("piece-cid"))
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		fmt.Println("CarFileName: ", carPath)
		if ctx.Bool("data-cid") {
			fmt.Println("DataCID: ", dataCid)
		}
		if ctx.Bool("piece-cid") {
			fmt.Println("PieceCID: ", pieceCid)
			fmt.Println("PieceSize: ", SizeStr(NewInt(pieceSize)))
			fmt.Println("Piece size in bytes: ", NewInt(pieceSize))
		}
		return nil
	},
}

var rpcApiCmd = &cli.Command{
	Name:      "rpc-api",
	Usage:     "RPC api proxy client of public chain",
	ArgsUsage: "[inputPath]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "chain-id",
			Usage: "chainId as public chain.",
		},
		&cli.StringFlag{
			Name:    "params",
			Aliases: []string{"p"},
			Usage:   "the parameters of the request api must be in string json format.",
		},
	},
	Action: func(ctx *cli.Context) error {
		chainId := ctx.String("chain-id")
		params := ctx.String("params")

		if utils.IsStrEmpty(&chainId) && utils.IsStrEmpty(&params) {
			err := fmt.Errorf("both chain-id and params are required")
			logs.GetLogger().Error(err)
			return ShowHelp(ctx, err)
		}
		result, err := command.SendRpcReqAndResp(chainId, params)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		fmt.Println("out:")
		print(string(result))
		return nil
	},
}

var rpcCmd = &cli.Command{
	Name:        "rpc",
	Usage:       "RPC proxy client of public chain",
	Subcommands: []*cli.Command{rpcLatestBalanceCmd, rpcCurrentHeightCmd},
}

var rpcCurrentHeightCmd = &cli.Command{
	Name:  "height",
	Usage: "Query current height of public chain",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "chain",
			Aliases: []string{"c"},
			Usage:   "public chain. support ETH、BNB、AVAX、MATIC、FTM、xDAI、IOTX、ONE、BOBA、FUSE、JEWEL、EVMOS、TUS",
		},
	},
	Action: func(ctx *cli.Context) error {
		chain := ctx.String("chain")
		result, err := command.QueryHeight(chain)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		fmt.Printf("Chain: %s\n", chain)
		fmt.Printf("Height: %d \n", result.Height)
		return nil
	},
}

var rpcLatestBalanceCmd = &cli.Command{
	Name:  "balance",
	Usage: "Query current balance of public chain",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "chain",
			Aliases: []string{"c"},
			Usage:   "public chain. support ETH、BNB、AVAX、MATIC、FTM、xDAI、IOTX、ONE、BOBA、FUSE、JEWEL、EVMOS、TUS",
		},
		&cli.StringFlag{
			Name:    "address",
			Aliases: []string{"a"},
			Usage:   "wallet address",
		},
	},
	Action: func(ctx *cli.Context) error {
		chain := ctx.String("chain")
		address := ctx.String("address")

		if utils.IsStrEmpty(&chain) && utils.IsStrEmpty(&chain) {
			err := errors.New("chain is required")
			logs.GetLogger().Error(err)
			return ShowHelp(ctx, err)
		}

		result, err := command.QueryChainInfo(chain, 0, address)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		fmt.Printf("Chain: %s\n", chain)
		fmt.Printf("Height: %d \n", result.Height)
		fmt.Printf("Address: %s \n", result.Address)
		fmt.Printf("Balance: %v \n", result.Balance)

		return nil
	},
}

func checkRepo() error {
	if strings.TrimSpace(config.GetConfig().Main.MarketVersion) == constants.MARKET_VERSION_2 {
		repoPath := config.GetConfig().Main.SwanRepo
		wallet := config.GetConfig().Sender.Wallet
		if _, err := os.Stat(repoPath); err != nil {
			if err := boost.GetClient(repoPath).InitRepo(repoPath, wallet); err != nil {
				return err
			}
		}
	}
	return nil
}

type BigInt = big2.Int

func NewInt(i uint64) BigInt {
	return BigInt{Int: big.NewInt(0).SetUint64(i)}
}

var byteSizeUnits = []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB"}

func SizeStr(bi big2.Int) string {
	r := new(big.Rat).SetInt(bi.Int)
	den := big.NewRat(1, 1024)

	var i int
	for f, _ := r.Float64(); f >= 1024 && i+1 < len(byteSizeUnits); f, _ = r.Float64() {
		i++
		r = r.Mul(r, den)
	}

	f, _ := r.Float64()
	return fmt.Sprintf("%.4g %s", f, byteSizeUnits[i])
}

type PrintHelpErr struct {
	Err error
	Ctx *cli.Context
}

func (e *PrintHelpErr) Error() string {
	return e.Err.Error()
}

func (e *PrintHelpErr) Unwrap() error {
	return e.Err
}

func (e *PrintHelpErr) Is(o error) bool {
	_, ok := o.(*PrintHelpErr)
	return ok
}

func ShowHelp(cctx *cli.Context, err error) error {
	return &PrintHelpErr{Err: err, Ctx: cctx}
}
