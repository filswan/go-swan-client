package test

import (
	"go-swan-client/common/client"
	"go-swan-client/logs"
	"go-swan-client/model"
	"go-swan-client/subcommand"
)

func TestGenerateCarFiles() {
	inputDir := "/home/peware/go-swan-client/input"
	outputDir := "/home/peware/go-swan-client/output"
	subcommand.GenerateCarFiles(&inputDir, &outputDir)
}

func TestCreateTask() {
	minerId := "miner_test"
	task := model.Task{
		TaskName:       "task_dora_test",
		CuratedDataset: "dataset",
		Description:    "description",
		IsPublic:       true,
		IsVerified:     true,
		MinerId:        &minerId,
	}

	swan := client.SwanGetClient()

	response := swan.SwanCreateTask(task, "/Users/dorachen/go-workspace/src/go-swan-client/test/car.csv")
	logs.GetLogger().Info(response)
}
