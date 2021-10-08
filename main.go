package main

import (
	"flag"
	"go-swan-client/logs"
	"go-swan-client/subcommand"
	"os"
)

const SUBCOMMAND_GENERATE_CAR = "car"
const SUBCOMMAND_GENERATE_GOCAR = "gocar"
const SUBCOMMAND_UPLOAD = "upload"
const SUBCOMMAND_CREATE_TASK = "task"

func main() {
	execSubCmd()
}

func execSubCmd() bool {
	if len(os.Args) < 2 {
		logs.GetLogger().Fatal("Sub command is required.")
	}

	subCmd := os.Args[1]
	switch subCmd {
	case SUBCOMMAND_GENERATE_CAR:
	case SUBCOMMAND_GENERATE_GOCAR:
		createCarFile(subCmd)
	case SUBCOMMAND_UPLOAD:
		uploadFile()
	case SUBCOMMAND_CREATE_TASK:
		createTask()
	default:
		logs.GetLogger().Error("Sub command should be: car|gocar|upload|task")
		return false
	}

	return true
}

//python3 swan_cli.py car --input-dir /home/peware/testGoSwanProvider/input --out-dir /home/peware/testGoSwanProvider/output
//go-swan-client car -input-dir ~/go-workspace/input/ -out-dir ~/go-workspace/output/
func createCarFile(subCmd string) bool {
	cmd := flag.NewFlagSet("car", flag.ExitOnError)

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
	case SUBCOMMAND_GENERATE_CAR:
		subcommand.GenerateCarFiles(inputDir, outputDir)
	case SUBCOMMAND_GENERATE_GOCAR:
		subcommand.GenerateGoCarFiles(inputDir, outputDir)
	default:
		return false
	}

	return true
}

//python3 swan_cli.py upload --input-dir /home/peware/testGoSwanProvider/output
func uploadFile() bool {
	cmd := flag.NewFlagSet("upload", flag.ExitOnError)

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
	cmd := flag.NewFlagSet("task", flag.ExitOnError)

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
