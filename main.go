package main

import (
	"encoding/json"
	"fmt"
	"github.com/filswan/go-swan-client/command"
	"github.com/filswan/go-swan-lib/logs"
	"github.com/filswan/go-swan-lib/utils"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
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
		Commands:        []*cli.Command{daemonCmd, toolsCmd, uploadCmd, taskCmd, dealCmd, autoCmd, rpcCmd},
		HideHelpCommand: true,
	}
	if err := app.Run(os.Args); err != nil {
		var phe *PrintHelpErr
		fmt.Fprintf(os.Stderr, "ERROR: %s\n\n", err)
		if xerrors.As(err, &phe) {
			_ = cli.ShowCommandHelp(phe.Ctx, phe.Ctx.Command.Name)
		}
		os.Exit(1)
	}
}

var daemonCmd = &cli.Command{
	Name:  "daemon",
	Usage: "Start a API service process",
	Action: func(ctx *cli.Context) error {
		router := httprouter.New()
		router.POST("/chain/rpc", func(writer http.ResponseWriter, request *http.Request, _ httprouter.Params) {
			defer request.Body.Close()
			data, _ := ioutil.ReadAll(request.Body)
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
	Usage:     "Upload CAR file to server",
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
	Usage: "Send deal task to swan",
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
			Name:  "auto",
			Usage: "automatically send the deal after the task is created",
			Value: false,
		},
		&cli.BoolFlag{
			Name:  "manual",
			Usage: "manually send the deal after the task is created",
			Value: false,
		},
		&cli.StringFlag{
			Name:  "miner",
			Usage: "target miner ID",
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
			Usage:   "max copy number you want to send",
			Value:   8,
		},
	},
	Action: func(ctx *cli.Context) error {
		inputDir := ctx.String("input-dir")
		if inputDir == "" {
			return errors.New("input-dir is required")
		}
		if !strings.HasSuffix(inputDir, "csv") && !strings.HasSuffix(inputDir, "json") {
			return errors.New("inputDir must be json or csv format file")
		}
		logs.GetLogger().Info("your input source file as: ", inputDir)

		auto := ctx.Bool("auto")
		manual := ctx.Bool("manual")
		minerId := ctx.String("miner")

		if auto && minerId != "" {
			return errors.New("miner cannot set when auto value is true")
		}

		if manual && minerId != "" {
			return errors.New("miner cannot set when manual value is true")
		}

		if !auto && !manual && minerId == "" {
			return errors.New("auto, manual, miner have at least one setting value")
		}

		outputDir := ctx.String("out-dir")
		_, fileDesc, _, total, err := command.CreateTaskByConfig(inputDir, &outputDir, ctx.String("name"), ctx.String("miner"),
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
	Usage: "Send auto bid deal",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "csv",
			Usage: "the CSV file path of deal metadata.",
		},
		&cli.StringFlag{
			Name:  "json",
			Usage: "the JSON file path of deal metadata.",
		},
		&cli.StringFlag{
			Name:    "out-dir",
			Aliases: []string{"o"},
			Usage:   "directory where target files will in.",
		},
		&cli.StringFlag{
			Name:  "miner",
			Usage: "target miner ID",
		},
	},
	Action: func(ctx *cli.Context) error {
		metadataJsonPath := ctx.String("json")
		metadataCsvPath := ctx.String("csv")

		if len(metadataJsonPath) == 0 && len(metadataCsvPath) == 0 {
			return errors.New("both metadataJsonPath and metadataCsvPath is nil")
		}

		if len(metadataJsonPath) > 0 && len(metadataCsvPath) > 0 {
			return errors.New("metadata file path is required, it cannot contain csv file path  or json file path  at the same time")
		}

		if len(metadataJsonPath) > 0 {
			logs.GetLogger().Info("Metadata json file:", metadataJsonPath)
		}
		if len(metadataCsvPath) > 0 {
			logs.GetLogger().Info("Metadata csv file:", metadataCsvPath)
		}

		_, err := command.SendDealsByConfig(ctx.String("out-dir"), ctx.String("miner"), metadataJsonPath, metadataCsvPath)
		if err != nil {
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
		err := command.SendAutoBidDealsLoopByConfig(ctx.String("out-dir"))
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		return nil
	},
}

var rpcCmd = &cli.Command{
	Name:      "rpc",
	Usage:     "RPC proxy client of public chain",
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
			return errors.New("both chain-id and params are required")
		}
		result, err := command.SendRpcReqAndResp(chainId, params)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		fmt.Println("out:")
		print(result)
		return nil
	},
}

var inPutFlag = cli.StringFlag{
	Name:     "input-dir",
	Aliases:  []string{"i"},
	Usage:    "directory where source file(s) is(are) in.",
	Required: true,
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
	},
	Action: func(ctx *cli.Context) error {
		inputDir := ctx.String("input-dir")
		if inputDir == "" {
			return errors.New("input-dir is required")
		}
		outputDir := ctx.String("out-dir")
		if _, err := command.CreateCarFilesByConfig(inputDir, &outputDir); err != nil {
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
		&cli.IntFlag{
			Name:  "parallel",
			Usage: "number goroutines run when building ipld nodes",
			Value: 5,
		},
		&cli.Int64Flag{
			Name:    "slice-size",
			Aliases: []string{"size"},
			Usage:   "size for each pieces",
			Value:   8589934592,
		},
		&cli.BoolFlag{
			Name:  "parent-path",
			Usage: "specify graph parent path",
			Value: false,
		},
	},
	Action: func(ctx *cli.Context) error {
		inputDir := ctx.String("input-dir")
		if inputDir == "" {
			return errors.New("input-dir is required")
		}
		outputDir := ctx.String("out-dir")
		if _, err := command.CreateGoCarFilesByConfig(inputDir, &outputDir, ctx.Int("parallel"), ctx.Int64("slice-size"), ctx.Bool("parent-path")); err != nil {
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
			Usage:    "specify source car path, directory or file",
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

var ipfsCarCmd = &cli.Command{
	Name:      "ipfs",
	Usage:     "Use ipfs api to generate car file",
	ArgsUsage: "[inputPath]",
	Flags: []cli.Flag{
		&inPutFlag,
		&outPutFlag,
	},
	Action: func(ctx *cli.Context) error {
		inputDir := ctx.String("input-dir")
		if inputDir == "" {
			return errors.New("input-dir is required")
		}
		outputDir := ctx.String("out-dir")
		if _, err := command.CreateIpfsCarFilesByConfig(inputDir, &outputDir); err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		return nil
	},
}

var toolsCmd = &cli.Command{
	Name:            "generate-car",
	Usage:           "Provide some very useful tools. (go-graphsplit, lotus, ipfs)",
	Subcommands:     []*cli.Command{splitCarCmd, lotusCarCmd, ipfsCarCmd},
	HideHelpCommand: true,
}

//var ipfsCmdCarCmd = &cli.Command{
//	Name:      "ipfs-cmd",
//	Usage:     "use the ipfs command to generate the car file",
//	ArgsUsage: "[inputPath]",
//	Flags: []cli.Flag{
//		&cli.StringFlag{
//			Name:    "input-dir",
//			Aliases: []string{"i"},
//			Usage:   "Directory where source file(s) is(are) in. (required)",
//		},
//		&cli.StringFlag{
//			Name:    "out-dir",
//			Aliases: []string{"o"},
//			Usage:   "Directory where car file(s) will be generated. (default)",
//		},
//	},
//	Action: func(ctx *cli.Context) error {
//		inputDir := ctx.String("input-dir")
//		if inputDir == "" {
//			return errors.New("input-dir is required")
//		}
//		outputDir := ctx.String("out-dir")
//		if _, err := command.CreateIpfsCmdCarFilesByConfig(inputDir, &outputDir); err != nil {
//			logs.GetLogger().Error(err)
//			return err
//		}
//		return nil
//	},
//}

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

func WithCategory(cat string, cmd *cli.Command) *cli.Command {
	cmd.Category = strings.ToUpper(cat)
	return cmd
}
