package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/DoraNebula/go-swan-client/logs"
	"github.com/DoraNebula/go-swan-client/subcommand"
)

const SUBCOMMAND_CAR = "car"
const SUBCOMMAND_GOCAR = "gocar"
const SUBCOMMAND_UPLOAD = "upload"
const SUBCOMMAND_TASK = "task"
const SUBCOMMAND_DEAL = "deal"
const SUBCOMMAND_AUTO = "auto"

func main() {
	execSubCmd()
	//subcommand.GoCar("",)
	//logs.GetLogger().Info("Hello")
	//test.Test()
}

func execSubCmd() error {
	if len(os.Args) < 2 {
		logs.GetLogger().Fatal("Sub command is required.")
	}

	var err error
	subCmd := os.Args[1]
	switch subCmd {
	case SUBCOMMAND_CAR, SUBCOMMAND_GOCAR:
		err = createCarFile(subCmd)
	case SUBCOMMAND_UPLOAD:
		err = uploadFile()
	case SUBCOMMAND_TASK:
		createTask()
	case SUBCOMMAND_DEAL:
		sendDeal()
	case SUBCOMMAND_AUTO:
		sendAutoBidDeal()
	default:
		err = fmt.Errorf("sub command should be: car|gocar|upload|task|deal")
		logs.GetLogger().Error(err)
	}

	return err
}

//python3 swan_cli.py car --input-dir /home/peware/testGoSwanProvider/input --out-dir /home/peware/testGoSwanProvider/output
//go-swan-client car -input-dir ~/go-workspace/input/ -out-dir ~/go-workspace/output/
func createCarFile(subCmd string) error {
	cmd := flag.NewFlagSet(subCmd, flag.ExitOnError)

	inputDir := cmd.String("input-dir", "", "Directory where source file(s) is(are) in.")
	outputDir := cmd.String("out-dir", "", "Directory where car file(s) will be generated.")

	err := cmd.Parse(os.Args[2:])
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	if !cmd.Parsed() {
		err = fmt.Errorf("sub command parse failed")
		logs.GetLogger().Error(err)
		return err
	}

	if inputDir == nil || len(*inputDir) == 0 {
		err = fmt.Errorf("input-dir is required")
		logs.GetLogger().Error(err)
		return err
	}

	switch subCmd {
	case SUBCOMMAND_CAR:
		outputDir, carFiles, err := subcommand.GenerateCarFiles(*inputDir, outputDir)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		logs.GetLogger().Info(len(carFiles), " car files generated to directory:", *outputDir)
	case SUBCOMMAND_GOCAR:
		outputDir, carFiles, err := subcommand.CreateGoCarFiles(*inputDir, outputDir)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		logs.GetLogger().Info(len(carFiles), " gocar files generated to directory:", *outputDir)
	default:
		err := fmt.Errorf("unknown sub command:%s", subCmd)
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}

//python3 swan_cli.py upload --input-dir /home/peware/testGoSwanProvider/output
func uploadFile() error {
	cmd := flag.NewFlagSet(SUBCOMMAND_UPLOAD, flag.ExitOnError)

	inputDir := cmd.String("input-dir", "", "Directory where source files are in.")

	err := cmd.Parse(os.Args[2:])
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	if !cmd.Parsed() {
		err := fmt.Errorf("sub command parse failed")
		logs.GetLogger().Error(err)
		return err
	}

	if inputDir == nil || len(*inputDir) == 0 {
		err := fmt.Errorf("input-dir is required")
		logs.GetLogger().Error(err)
		return err
	}

	err = subcommand.UploadCarFiles(*inputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}

//python3 swan_cli.py task --input-dir /home/peware/testGoSwanProvider/output --out-dir /home/peware/testGoSwanProvider/task --miner t03354 --dataset test --description test
func createTask() error {
	cmd := flag.NewFlagSet(SUBCOMMAND_TASK, flag.ExitOnError)

	taskName := cmd.String("name", "", "Directory where source files are in.")
	inputDir := cmd.String("input-dir", "", "Directory where source files are in.")
	outputDir := cmd.String("out-dir", "", "Directory where target files will in.")
	minerFid := cmd.String("miner", "", "Target miner fid")
	dataset := cmd.String("dataset", "", "Curated dataset.")
	description := cmd.String("description", "", "Task description.")

	err := cmd.Parse(os.Args[2:])
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	if !cmd.Parsed() {
		err = fmt.Errorf("sub command parse failed")
		logs.GetLogger().Error(err)
		return err
	}

	if inputDir == nil || len(*inputDir) == 0 {
		err = fmt.Errorf("input-dir is required")
		logs.GetLogger().Error(err)
		return err
	}

	logs.GetLogger().Info(inputDir, outputDir, minerFid, dataset, description)

	subcommand.CreateTask(*inputDir, taskName, outputDir, minerFid, dataset, description)
	return nil
}

func sendDeal() bool {
	cmd := flag.NewFlagSet(SUBCOMMAND_DEAL, flag.ExitOnError)

	metadataJsonPath := cmd.String("json", "", "The JSON file path of deal metadata.")
	outputDir := cmd.String("out-dir", "", "Directory where target files will in.")
	minerFid := cmd.String("miner", "", "Target miner fid")

	err := cmd.Parse(os.Args[2:])
	if err != nil {
		logs.GetLogger().Error(err)
		return false
	}

	if !cmd.Parsed() {
		logs.GetLogger().Error("Sub command parse failed.")
		return false
	}

	if metadataJsonPath == nil || len(*metadataJsonPath) == 0 {
		logs.GetLogger().Error("input-dir is required.")
		return false
	}

	if minerFid == nil || len(*minerFid) == 0 {
		logs.GetLogger().Error("miner is required.")
		return false
	}

	logs.GetLogger().Info("metadata json file:", *metadataJsonPath)
	logs.GetLogger().Info("output dir:", *outputDir)
	logs.GetLogger().Info("miner:", *minerFid)

	result := subcommand.SendDeals(*minerFid, outputDir, *metadataJsonPath)
	return result
}

func sendAutoBidDeal() bool {
	cmd := flag.NewFlagSet(SUBCOMMAND_DEAL, flag.ExitOnError)

	outputDir := cmd.String("out-dir", "", "Directory where target files will in.")

	err := cmd.Parse(os.Args[2:])
	if err != nil {
		logs.GetLogger().Error(err)
		return false
	}

	if !cmd.Parsed() {
		logs.GetLogger().Error("Sub command parse failed.")
		return false
	}

	subcommand.SendAutoBidDeal(outputDir)

	return true
}
