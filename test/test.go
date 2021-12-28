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
	//TestCreateIpfsCarFiles()
	//TestUpload()
	//TestCreateTask()
	//TestSendDeals()
	//TestSendAutoBidDeals()
	TestSendAutoBidDealsByTaskUuid()
}

func estCreateCarFiles() {
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

	command.CreateTaskByConfig(inputDir, &outDir, "", "t03354", "", "")
}

func TestSendDeals() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}

	outDir := filepath.Join(homeDir, "work/carFiles")

	command.SendDealsByConfig(outDir, "t03354", "/Users/dorachen/work/carFiles/swan-task-eqzgmt-metadata.json")
}

func TestSendAutoBidDeals() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}

	outDir := filepath.Join(homeDir, "work/carFiles")

	command.SendAutoBidDealsLoopByConfig(outDir)
}

func TestSendAutoBidDealsByTaskUuid() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}

	outDir := filepath.Join(homeDir, "work/carFiles")

	cmdAutoBidDeal := command.GetCmdAutoDeal(&outDir)
	cmdAutoBidDeal.SendAutoBidDealsByTaskUuid("a6068405-2cc0-40d1-bf22-8dffbcb17060")
}
