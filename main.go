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
	//test.TestGenerateCarFiles()
	execCmd()
}

func execCmd() {
	if len(os.Args) < 2 {
		logs.GetLogger().Fatal("Sub command is required.")
	}

	switch os.Args[1] {
	case SUBCOMMAND_GENERATE_CAR:
		inputDir, outputDir := getCarArgs()
		subcommand.GenerateCarFiles(inputDir, outputDir)
	case SUBCOMMAND_GENERATE_GOCAR:
		inputDir, outputDir := getCarArgs()
		subcommand.GenerateGoCarFiles(inputDir, outputDir)
	case SUBCOMMAND_UPLOAD:
		inputDir := getUploadArgs()
		subcommand.UploadCarFiles(*inputDir)
	case SUBCOMMAND_CREATE_TASK:
		taskName, inputDir, outputDir, minerFid, dataset, description := getTaskArgs()
		subcommand.CreateTask(taskName, inputDir, outputDir, minerFid, dataset, description)
	default:
		logs.GetLogger().Fatal("Sub command should be: car|gocar|upload|task")
	}
}

//python3 swan_cli.py car --input-dir /home/peware/testGoSwanProvider/input --out-dir /home/peware/testGoSwanProvider/output
func getCarArgs() (*string, *string) {
	cmd := flag.NewFlagSet("car", flag.ExitOnError)

	inputDir := cmd.String("input-dir", "", "Directory where source file(s) is(are) in.")
	outputDir := cmd.String("output-dir", "", "Directory where car file(s) will be generated.")

	err := cmd.Parse(os.Args[2:])
	if err != nil {
		logs.GetLogger().Fatal(err)
	}

	if !cmd.Parsed() {
		logs.GetLogger().Fatal("Sub command parse failed.")
	}

	if inputDir == nil || len(*inputDir) == 0 {
		logs.GetLogger().Fatal("input-dir is required.")
	}

	return inputDir, outputDir
}

//python3 swan_cli.py upload --input-dir /home/peware/testGoSwanProvider/output
func getUploadArgs() *string {
	cmd := flag.NewFlagSet("upload", flag.ExitOnError)

	inputDir := cmd.String("input-dir", "", "Directory where source files are in.")

	err := cmd.Parse(os.Args[2:])
	if err != nil {
		logs.GetLogger().Fatal(err)
	}

	if !cmd.Parsed() {
		logs.GetLogger().Fatal("Sub command parse failed.")
	}

	if inputDir == nil || len(*inputDir) == 0 {
		logs.GetLogger().Fatal("input-dir is required.")
	}

	return inputDir
}

//python3 swan_cli.py task --input-dir /home/peware/testGoSwanProvider/output --out-dir /home/peware/testGoSwanProvider/task --miner t03354 --dataset test --description test
func getTaskArgs() (*string, *string, *string, *string, *string, *string) {
	cmd := flag.NewFlagSet("task", flag.ExitOnError)

	taskName := cmd.String("name", "", "Directory where source files are in.")
	inputDir := cmd.String("input-dir", "", "Directory where source files are in.")
	outputDir := cmd.String("output-dir", "", "Directory where target files will in.")
	minerFid := cmd.String("miner", "", "Target miner fid")
	dataset := cmd.String("dataset", "", "Curated dataset.")
	description := cmd.String("description", "", "Task description.")

	err := cmd.Parse(os.Args[2:])
	if err != nil {
		logs.GetLogger().Fatal(err)
	}

	if !cmd.Parsed() {
		logs.GetLogger().Fatal("Sub command parse failed.")
	}

	if inputDir == nil || len(*inputDir) == 0 {
		logs.GetLogger().Fatal("input-dir is required.")
	}

	logs.GetLogger().Info(inputDir, outputDir, minerFid, dataset, description)
	return taskName, inputDir, outputDir, minerFid, dataset, description
}
