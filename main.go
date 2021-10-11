package main

import (
	"flag"
	"go-swan-client/logs"
	"go-swan-client/subcommand"
	"os"
)

const SUBCOMMAND_CAR = "car"
const SUBCOMMAND_GOCAR = "gocar"
const SUBCOMMAND_UPLOAD = "upload"
const SUBCOMMAND_TASK = "task"
const SUBCOMMAND_DEAL = "deal"

func main() {
	execSubCmd()
	//logs.GetLogger().Info("Hello")
	//test.Test()
}

func execSubCmd() bool {
	if len(os.Args) < 2 {
		logs.GetLogger().Fatal("Sub command is required.")
	}

	result := true
	subCmd := os.Args[1]
	switch subCmd {
	case SUBCOMMAND_CAR, SUBCOMMAND_GOCAR:
		result = createCarFile(subCmd)
	case SUBCOMMAND_UPLOAD:
		result = uploadFile()
	case SUBCOMMAND_TASK:
		result = createTask()
	case SUBCOMMAND_DEAL:
		result = sendDeal()
	default:
		logs.GetLogger().Error("Sub command should be: car|gocar|upload|task")
		result = false
	}

	return result
}

//python3 swan_cli.py car --input-dir /home/peware/testGoSwanProvider/input --out-dir /home/peware/testGoSwanProvider/output
//go-swan-client car -input-dir ~/go-workspace/input/ -out-dir ~/go-workspace/output/
func createCarFile(subCmd string) bool {
	cmd := flag.NewFlagSet(subCmd, flag.ExitOnError)

	inputDir := cmd.String("input-dir", "", "Directory where source file(s) is(are) in.")
	outputDir := cmd.String("out-dir", "", "Directory where car file(s) will be generated.")

	err := cmd.Parse(os.Args[2:])
	if err != nil {
		logs.GetLogger().Error(err)
		return false
	}

	if !cmd.Parsed() {
		logs.GetLogger().Error("Sub command parse failed.")
		return false
	}

	if inputDir == nil || len(*inputDir) == 0 {
		logs.GetLogger().Error("input-dir is required.")
		return false
	}

	switch subCmd {
	case SUBCOMMAND_CAR:
		subcommand.GenerateCarFiles(inputDir, outputDir)
	case SUBCOMMAND_GOCAR:
		subcommand.GenerateGoCarFiles(inputDir, outputDir)
	default:
		return false
	}

	return true
}

//python3 swan_cli.py upload --input-dir /home/peware/testGoSwanProvider/output
func uploadFile() bool {
	cmd := flag.NewFlagSet(SUBCOMMAND_UPLOAD, flag.ExitOnError)

	inputDir := cmd.String("input-dir", "", "Directory where source files are in.")

	err := cmd.Parse(os.Args[2:])
	if err != nil {
		logs.GetLogger().Error(err)
		return false
	}

	if !cmd.Parsed() {
		logs.GetLogger().Error("Sub command parse failed.")
		return false
	}

	if inputDir == nil || len(*inputDir) == 0 {
		logs.GetLogger().Error("input-dir is required.")
		return false
	}

	subcommand.UploadCarFiles(*inputDir)

	return true
}

//python3 swan_cli.py task --input-dir /home/peware/testGoSwanProvider/output --out-dir /home/peware/testGoSwanProvider/task --miner t03354 --dataset test --description test
func createTask() bool {
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
		return false
	}

	if !cmd.Parsed() {
		logs.GetLogger().Error("Sub command parse failed.")
		return false
	}

	if inputDir == nil || len(*inputDir) == 0 {
		logs.GetLogger().Error("input-dir is required.")
		return false
	}

	logs.GetLogger().Info(inputDir, outputDir, minerFid, dataset, description)

	subcommand.CreateTask(taskName, inputDir, outputDir, minerFid, dataset, description)
	return true
}

func sendDeal() bool {
	cmd := flag.NewFlagSet(SUBCOMMAND_DEAL, flag.ExitOnError)

	metadataCsvPath := cmd.String("csv", "", "The CSV file path of deal metadata.")
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

	if metadataCsvPath == nil || len(*metadataCsvPath) == 0 {
		logs.GetLogger().Error("input-dir is required.")
		return false
	}

	if minerFid == nil || len(*minerFid) == 0 {
		logs.GetLogger().Error("input-dir is required.")
		return false
	}

	logs.GetLogger().Info(metadataCsvPath, outputDir, minerFid)

	//subcommand.CreateTask(taskName, inputDir, outputDir, minerFid, dataset, description)
	return true
}
