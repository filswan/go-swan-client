package test

import (
	"os"
	"path/filepath"

	"github.com/filswan/go-swan-client/subcommand"
	"github.com/filswan/go-swan-lib/logs"
)

func Test() {
	TestCreateTask()
}

func TestCreateTask() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}
	inputDir := filepath.Join(homeDir, "work/carFiles")
	outDir := filepath.Join(homeDir, "work/carFiles")

	subcommand.CreateTaskByConfig(inputDir, &outDir, "", "", "", "")
}
