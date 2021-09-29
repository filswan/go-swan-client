package main

import (
	"fmt"
	"go-swan-client/logs"
	"os"
	"strconv"
)

const OPERATION_TYPE_CAR = "car"
const OPERATION_TYPE_UPLOAD = "upload"
const OPERATION_TYPE_TASK = "task"

func main() {
	if len(os.Args) < 2 {
		logs.GetLogger().Error("Not enough arguments.")
		return
	}

	operationType := os.Args[1]

	switch operationType {
	case OPERATION_TYPE_CAR:
		GenerateCarFiles()
	case OPERATION_TYPE_UPLOAD:
		UploadFiles()
	case OPERATION_TYPE_TASK:
		CreateTask()
	default:
		logs.GetLogger().Error("Unknow operation type.")
		return
	}
	filepath := os.Args[1]
	filename := os.Args[2]
	filesizeInGigabyte, err := strconv.ParseInt(os.Args[3], 10, 64)
	if err != nil {
		fmt.Println(err)
	}

	logs.GetLogger().Info(filepath, filename, filesizeInGigabyte)
}

func GenerateCarFiles() {
	//python3 swan_cli.py car --input-dir /home/peware/testGoSwanProvider/input --out-dir /home/peware/testGoSwanProvider/output
	inputDir := os.Args[2]
	outputDir := os.Args[3]

	logs.GetLogger().Info(inputDir, outputDir)
}

func UploadFiles() {
	//python3 swan_cli.py upload --input-dir /home/peware/testGoSwanProvider/output
	inputDir := os.Args[2]

	logs.GetLogger().Info(inputDir)
}

func CreateTask() {
	//python3 swan_cli.py task --input-dir /home/peware/testGoSwanProvider/output --out-dir /home/peware/testGoSwanProvider/task --miner t03354 --dataset test --description test
	inputDir := os.Args[2]
	outputDir := os.Args[3]
	minerFid := os.Args[4]

}
