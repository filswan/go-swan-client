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
	//TestCreateTasks(1)
	//TestSendDeals()
	TestSendAutoBidDeals()
	//TestSendAutoBidDealsByTaskUuid()
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

func TestCreateTasks(taskCnt int) {
	for i := 0; i < taskCnt; i++ {
		TestCreateTask()
	}
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

	command.SendDealsByConfig(outDir, "t03354", "", "/Users/dorachen/work/carFiles/swan-task-oe1p20-metadata.json")
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
	jsonFilepath, fileDescs, err := cmdAutoBidDeal.SendAutoBidDealsByTaskUuid("e04bd920-bab4-498a-afb9-8f9a4222c895")
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}
	logs.GetLogger().Info(*jsonFilepath, *fileDescs[0])
}
