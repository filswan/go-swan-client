package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/filswan/go-swan-client/command"

	"github.com/filswan/go-swan-lib/logs"
)

func main() {
	execSubCmd()
	//test.Test()
}

func execSubCmd() error {
	if len(os.Args) < 2 {
		logs.GetLogger().Fatal("Sub command is required.")
	}

	var err error = nil
	subCmd := os.Args[1]
	switch subCmd {
	case command.CMD_CAR, command.CMD_GOCAR, command.CMD_IPFSCAR, command.CMD_IPFSCMDCAR:
		err = createCarFile(subCmd)
	case command.CMD_UPLOAD:
		err = uploadFile()
	case command.CMD_TASK:
		err = createTask()
	case command.CMD_DEAL:
		err = sendDeal()
	case command.CMD_AUTO:
		err = sendAutoBidDeal()
	default:
		err = fmt.Errorf("sub command should be: car|gocar|upload|task|deal|auto")
		logs.GetLogger().Error(err)
	}

	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}

func createCarFile(subCmd string) error {
	cmd := flag.NewFlagSet(subCmd, flag.ExitOnError)

	inputDir := cmd.String("input-dir", "", "Directory where source file(s) is(are) in.")
	outputDir := cmd.String("out-dir", "", "Directory where car file(s) will be generated.")

	err := cmd.Parse(os.Args[2:])
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	if !cmd.Parsed() {
		err = fmt.Errorf("sub command parse failed")
		logs.GetLogger().Error(err)
		return err
	}

	if inputDir == nil || len(*inputDir) == 0 {
		err = fmt.Errorf("input-dir is required")
		logs.GetLogger().Error(err)
		return err
	}

	switch subCmd {
	case command.CMD_CAR:
		_, err = command.CreateCarFilesByConfig(*inputDir, outputDir)
	case command.CMD_GOCAR:
		_, err = command.CreateGoCarFilesByConfig(*inputDir, outputDir)
	case command.CMD_IPFSCAR:
		_, err = command.CreateIpfsCarFilesByConfig(*inputDir, outputDir)
	case command.CMD_IPFSCMDCAR:
		_, err = command.CreateIpfsCmdCarFilesByConfig(*inputDir, outputDir)
	default:
		err = fmt.Errorf("unknown sub command:%s", subCmd)
	}

	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}

func uploadFile() error {
	cmd := flag.NewFlagSet(command.CMD_UPLOAD, flag.ExitOnError)

	inputDir := cmd.String("input-dir", "", "Directory where source files are in.")

	err := cmd.Parse(os.Args[2:])
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	if !cmd.Parsed() {
		err := fmt.Errorf("sub command parse failed")
		logs.GetLogger().Error(err)
		return err
	}

	if inputDir == nil || len(*inputDir) == 0 {
		err := fmt.Errorf("input-dir is required")
		logs.GetLogger().Error(err)
		return err
	}

	_, err = command.UploadCarFilesByConfig(*inputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}

func createTask() error {
	cmd := flag.NewFlagSet(command.CMD_TASK, flag.ExitOnError)

	taskName := cmd.String("name", "", "Directory where source files are in.")
	inputDir := cmd.String("input-dir", "", "Directory where source files are in.")
	outputDir := cmd.String("out-dir", "", "Directory where target files will in.")
	minerFid := cmd.String("miner", "", "Target miner fid")
	dataset := cmd.String("dataset", "", "Curated dataset.")
	description := cmd.String("description", "", "Task description.")

	err := cmd.Parse(os.Args[2:])
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	if !cmd.Parsed() {
		err = fmt.Errorf("sub command parse failed")
		logs.GetLogger().Error(err)
		return err
	}

	if inputDir == nil || len(*inputDir) == 0 {
		err = fmt.Errorf("input-dir is required")
		logs.GetLogger().Error(err)
		return err
	}

	logs.GetLogger().Info("your input dir: ", *inputDir)

	_, _, _, err = command.CreateTaskByConfig(*inputDir, outputDir, *taskName, *minerFid, *dataset, *description)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}

func sendDeal() error {
	cmd := flag.NewFlagSet(command.CMD_DEAL, flag.ExitOnError)

	metadataJsonPath := cmd.String("json", "", "The JSON file path of deal metadata.")
	outputDir := cmd.String("out-dir", "", "Directory where target files will in.")
	minerFid := cmd.String("miner", "", "Target miner fid")

	err := cmd.Parse(os.Args[2:])
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	if !cmd.Parsed() {
		err := fmt.Errorf("sub command parse failed")
		logs.GetLogger().Error(err)
		return err
	}

	if metadataJsonPath == nil || len(*metadataJsonPath) == 0 {
		err := fmt.Errorf("metadata json file path is required")
		logs.GetLogger().Error(err)
		return err
	}

	logs.GetLogger().Info("Metadata json file:", *metadataJsonPath)
	logs.GetLogger().Info("Output dir:", *outputDir)

	_, err = command.SendDealsByConfig(*outputDir, *minerFid, *metadataJsonPath)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}

func sendAutoBidDeal() error {
	cmd := flag.NewFlagSet(command.CMD_DEAL, flag.ExitOnError)

	outputDir := cmd.String("out-dir", "", "Directory where target files will in.")

	err := cmd.Parse(os.Args[2:])
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	if !cmd.Parsed() {
		err := fmt.Errorf("sub command parse failed")
		logs.GetLogger().Error(err)
		return err
	}

	command.SendAutoBidDealsLoopByConfig(*outputDir)
	return nil
}
