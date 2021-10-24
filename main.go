package main

import (
	"flag"
	"fmt"
	"os"

	"go-swan-client/logs"
	"go-swan-client/model"
	"go-swan-client/subcommand"
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

	var err error = nil
	subCmd := os.Args[1]
	switch subCmd {
	case SUBCOMMAND_CAR, SUBCOMMAND_GOCAR:
		err = createCarFile(subCmd)
	case SUBCOMMAND_UPLOAD:
		err = uploadFile()
	case SUBCOMMAND_TASK:
		err = createTask()
	case SUBCOMMAND_DEAL:
		err = sendDeal()
	case SUBCOMMAND_AUTO:
		err = sendAutoBidDeal()
	default:
		err = fmt.Errorf("sub command should be: car|gocar|upload|task|deal|auto")
		logs.GetLogger().Error(err)
	}

	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	return nil
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

	confCar := model.GetConfCar(outputDir)

	switch subCmd {
	case SUBCOMMAND_CAR:
		_, err := subcommand.GenerateCarFiles(confCar, *inputDir)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		//logs.GetLogger().Info(len(carFiles), " car files generated to directory:", *outputDir)
	case SUBCOMMAND_GOCAR:
		_, err := subcommand.CreateGoCarFiles(confCar, *inputDir)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		//logs.GetLogger().Info(len(carFiles), " gocar files generated to directory:", *outputDir)
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

	confUpload := model.GetConfUpload()

	err = subcommand.UploadCarFiles(confUpload, *inputDir)
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

	logs.GetLogger().Info("your input dir: ", *inputDir)

	confTask := model.GetConfTask(outputDir)
	jsonFileName, err := subcommand.CreateTask(confTask, *inputDir, taskName, minerFid, dataset, description)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	logs.GetLogger().Info("Task information is in:", *jsonFileName)

	return nil
}

func sendDeal() error {
	cmd := flag.NewFlagSet(SUBCOMMAND_DEAL, flag.ExitOnError)

	metadataJsonPath := cmd.String("json", "", "The JSON file path of deal metadata.")
	outputDir := cmd.String("out-dir", "", "Directory where target files will in.")
	minerFid := cmd.String("miner", "", "Target miner fid")

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

	if metadataJsonPath == nil || len(*metadataJsonPath) == 0 {
		err := fmt.Errorf("json is required")
		logs.GetLogger().Error(err)
		return err
	}

	if minerFid == nil || len(*minerFid) == 0 {
		err := fmt.Errorf("miner is required")
		logs.GetLogger().Error(err)
		return err
	}

	logs.GetLogger().Info("metadata json file:", *metadataJsonPath)
	logs.GetLogger().Info("output dir:", *outputDir)
	logs.GetLogger().Info("miner:", *minerFid)

	confDeal := model.GetConfDeal(outputDir, minerFid)
	err = subcommand.SendDeals(*confDeal, *minerFid, outputDir, *metadataJsonPath)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}

func sendAutoBidDeal() error {
	cmd := flag.NewFlagSet(SUBCOMMAND_DEAL, flag.ExitOnError)

	outputDir := cmd.String("out-dir", "", "Directory where target files will in.")

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

	confDeal := model.GetConfDeal(outputDir, nil)
	csvFilepaths, err := subcommand.SendAutoBidDeal(confDeal)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	for _, csvFilepath := range csvFilepaths {
		logs.GetLogger().Info(csvFilepath, " is generated")
	}

	return nil
}
