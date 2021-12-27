package test

import (
	"os"
	"path/filepath"

	"github.com/filswan/go-swan-client/command"
	"github.com/filswan/go-swan-lib/logs"
)

func Test() {
	//TestCreateCarFiles()
	//TestCreateGoCarFiles()
	TestCreateIpfsCarFiles()
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

	command.CreateCarFilesByConfig(inputDir, &outDir)
}

func TestCreateGoCarFiles() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}
	inputDir := filepath.Join(homeDir, "work/srcFiles")
	outDir := filepath.Join(homeDir, "work/carFiles")

	command.CreateGoCarFilesByConfig(inputDir, &outDir)
}

func TestCreateIpfsCarFiles() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}
	inputDir := filepath.Join(homeDir, "work/srcFiles")
	outDir := filepath.Join(homeDir, "work/carFiles")

	command.CreateIpfsCarFilesByConfig(inputDir, &outDir)
}

func TestUpload() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}
	inputDir := filepath.Join(homeDir, "work/carFiles")

	command.UploadCarFilesByConfig(inputDir)
}

func TestCreateTask() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}
	inputDir := filepath.Join(homeDir, "work/carFiles")
	outDir := filepath.Join(homeDir, "work/carFiles")

	command.CreateTaskByConfig(inputDir, &outDir, "", "", "", "")
}
