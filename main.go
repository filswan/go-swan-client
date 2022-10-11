package main

import (
	"encoding/json"
	"fmt"
	"github.com/filswan/go-swan-client/command"
	"github.com/filswan/go-swan-lib/logs"
	"github.com/filswan/go-swan-lib/utils"
	"github.com/julienschmidt/httprouter"
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
		Usage:                "A PiB level data ooborading tool for Filecoin Network",
		Version:              command.VERSION,
		EnableBashCompletion: true,
		After: func(context *cli.Context) error {
			if r := recover(); r != nil {
				panic(r)
			}
			return nil
		},
		Commands: []*cli.Command{
			daemonCmd, carCmd, uploadCmd, taskCmd, dealCmd, autoCmd, rpcCmd, VersionCmd},
	}
	if err := app.Run(os.Args); err != nil {
		var phe *PrintHelpErr
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

var VersionCmd = &cli.Command{
	Name:  "version",
	Usage: "Print version",
	Action: func(ctx *cli.Context) error {
		cli.VersionPrinter(ctx)
		return nil
	},
}

var uploadCmd = &cli.Command{
	Name:      "upload",
	Usage:     "upload car file to server",
	ArgsUsage: "[inputPath]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "input-dir",
			Aliases: []string{"i"},
			Usage:   "Directory where source files are in.",
		},
	},
	Action: func(ctx *cli.Context) error {
		if !ctx.Args().Present() {
			return cli.ShowCommandHelp(ctx, ctx.Command.Name)
		}
		inputDir := ctx.String("input-dir")
		if inputDir == "" {
			err := fmt.Errorf("input-dir is required")
			logs.GetLogger().Error(err)
			return ShowHelp(ctx, err)
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
	Usage: "send deal task to swan",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "name",
			Usage: "Directory where source files are in.",
		},
		&cli.StringFlag{
			Name:    "input-dir",
			Aliases: []string{"i"},
			Usage:   "Absolute path where the json or csv format source files.(required)",
		},
		&cli.StringFlag{
			Name:    "out-dir",
			Aliases: []string{"o"},
			Usage:   "Directory where target files will in.(default)",
		},
		&cli.StringFlag{
			Name:  "miner",
			Usage: "Target miner Id",
		},
		&cli.StringFlag{
			Name:  "dataset",
			Usage: "Curated dataset.",
		},
		&cli.StringFlag{
			Name:  "description",
			Usage: "Task description.",
		},
	},
	Action: func(ctx *cli.Context) error {
		if !ctx.Args().Present() {
			return cli.ShowCommandHelp(ctx, ctx.Command.Name)
		}
		inputDir := ctx.String("input-dir")
		if inputDir == "" {
			err := fmt.Errorf("input-dir is required")
			logs.GetLogger().Error(err)
			return ShowHelp(ctx, err)
		}
		if !strings.HasSuffix(inputDir, "csv") && !strings.HasSuffix(inputDir, "json") {
			err := fmt.Errorf("inputDir must be json or csv format file")
			logs.GetLogger().Error(err)
			return ShowHelp(ctx, err)
		}
		logs.GetLogger().Info("your input source file as: ", inputDir)
		outputDir := ctx.String("out-dir")
		_, _, _, err := command.CreateTaskByConfig(inputDir, &outputDir, ctx.String("name"), ctx.String("miner"), ctx.String("dataset"), ctx.String("description"))
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		return nil
	},
}

var dealCmd = &cli.Command{
	Name:  "deal",
	Usage: "send auto bid deal",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "csv",
			Usage: "The CSV file path of deal metadata.",
		},
		&cli.StringFlag{
			Name:  "json",
			Usage: "The JSON file path of deal metadata.",
		},
		&cli.StringFlag{
			Name:    "out-dir",
			Aliases: []string{"o"},
			Usage:   "Directory where target files will in.",
		},
		&cli.StringFlag{
			Name:  "miner",
			Usage: "Target miner fid",
		},
	},
	Action: func(ctx *cli.Context) error {
		if !ctx.Args().Present() {
			return cli.ShowCommandHelp(ctx, ctx.Command.Name)
		}

		metadataJsonPath := ctx.String("json")
		metadataCsvPath := ctx.String("csv")

		if len(metadataJsonPath) == 0 && len(metadataCsvPath) == 0 {
			err := fmt.Errorf("both metadataJsonPath and metadataCsvPath is nil")
			logs.GetLogger().Error(err)
			return ShowHelp(ctx, err)
		}

		if len(metadataJsonPath) > 0 && len(metadataCsvPath) > 0 {
			err := fmt.Errorf("metadata file path is required, it cannot contain csv file path  or json file path  at the same time")
			logs.GetLogger().Error(err)
			return ShowHelp(ctx, err)
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
	Usage:     "auto send bid deal",
	ArgsUsage: "[inputPath]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "out-dir",
			Aliases: []string{"o"},
			Usage:   "Directory where target files will in.",
		},
	},
	Action: func(ctx *cli.Context) error {
		if !ctx.Args().Present() {
			return cli.ShowCommandHelp(ctx, ctx.Command.Name)
		}
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
	Usage:     "rpc proxy client of public chain",
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
		if !ctx.Args().Present() {
			return cli.ShowCommandHelp(ctx, ctx.Command.Name)
		}
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
		print(result)
		return nil
	},
}

var carCmd = &cli.Command{
	Name:      "car",
	Usage:     "generate car file",
	ArgsUsage: "[inputPath]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "input-dir",
			Aliases: []string{"i"},
			Usage:   "Directory where source file(s) is(are) in.",
		},
		&cli.StringFlag{
			Name:    "out-dir",
			Aliases: []string{"o"},
			Usage:   "Directory where car file(s) will be generated.",
		},
		&cli.StringFlag{
			Name:  "mode",
			Usage: "support four mode: lotus,split,ipfs,ipfscmd",
		},
	},
	Action: func(ctx *cli.Context) error {
		if !ctx.Args().Present() {
			return cli.ShowCommandHelp(ctx, ctx.Command.Name)
		}
		inputDir := ctx.String("input-dir")
		if inputDir == "" {
			err := fmt.Errorf("input-dir is required")
			logs.GetLogger().Error(err)
			return err
		}
		outputDir := ctx.String("out-dir")
		mode := ctx.String("mode")
		if mode == "" {
			err := fmt.Errorf("mode is required")
			logs.GetLogger().Error(err)
			return err
		}
		var err error
		switch mode {
		case command.CMD_CAR:
			_, err = command.CreateCarFilesByConfig(inputDir, &outputDir)
		case command.CMD_GOCAR:
			_, err = command.CreateGoCarFilesByConfig(inputDir, &outputDir)
		case command.CMD_IPFSCAR:
			_, err = command.CreateIpfsCarFilesByConfig(inputDir, &outputDir)
		case command.CMD_IPFSCMDCAR:
			_, err = command.CreateIpfsCmdCarFilesByConfig(inputDir, &outputDir)
		default:
			err = fmt.Errorf("not support mode:%s, only support: car,gocar,ipfscar,ipfscmdcar", mode)
		}
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		return nil
	},
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
