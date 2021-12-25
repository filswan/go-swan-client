package test

import (
	"os"
	"path/filepath"

	"github.com/filswan/go-swan-client/subcommand"
	"github.com/filswan/go-swan-lib/logs"
)

func Test() {
	TestCreateCarFiles()
	//TestCreateTask()
}

func TestCreateCarFiles() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}
	inputDir := filepath.Join(homeDir, "work/srcFiles")
	outDir := filepath.Join(homeDir, "work/carFiles")

	subcommand.CreateCarFilesByConfig(inputDir, &outDir)
}

func TestCreateGoCarFiles() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}
	inputDir := filepath.Join(homeDir, "work/srcFiles")
	outDir := filepath.Join(homeDir, "work/carFiles")

	subcommand.CreateCarFilesByConfig(inputDir, &outDir)
}

func TestCreateIpfsCarFiles() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}
	inputDir := filepath.Join(homeDir, "work/srcFiles")
	outDir := filepath.Join(homeDir, "work/carFiles")

	subcommand.CreateCarFilesByConfig(inputDir, &outDir)
}

func TestUpload() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}
	inputDir := filepath.Join(homeDir, "work/carFiles")

	subcommand.UploadCarFilesByConfig(inputDir)
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
