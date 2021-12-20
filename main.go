package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/filswan/go-swan-client/model"

	"github.com/filswan/go-swan-client/subcommand"

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
	case subcommand.SUBCOMMAND_CAR, subcommand.SUBCOMMAND_GOCAR, subcommand.SUBCOMMAND_IPFSCAR:
		err = createCarFile(subCmd)
	case subcommand.SUBCOMMAND_UPLOAD:
		err = uploadFile()
	case subcommand.SUBCOMMAND_TASK:
		err = createTask()
	case subcommand.SUBCOMMAND_DEAL:
		err = sendDeal()
	case subcommand.SUBCOMMAND_AUTO:
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

	confCar := model.GetConfCar(*inputDir, outputDir)

	switch subCmd {
	case subcommand.SUBCOMMAND_CAR:
		_, err := subcommand.CreateCarFiles(confCar)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		//logs.GetLogger().Info(len(carFiles), " car files generated to directory:", *outputDir)
	case subcommand.SUBCOMMAND_GOCAR:
		_, err := subcommand.CreateGoCarFiles(confCar)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
	case subcommand.SUBCOMMAND_IPFSCAR:
		_, err := subcommand.CreateIpfsCarFiles(confCar)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
		//logs.GetLogger().Info(len(carFiles), " gocar files generated to directory:", *outputDir)
	default:
		err := fmt.Errorf("unknown sub command:%s", subCmd)
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}

//python3 swan_cli.py upload --input-dir /home/peware/testGoSwanProvider/output
func uploadFile() error {
	cmd := flag.NewFlagSet(subcommand.SUBCOMMAND_UPLOAD, flag.ExitOnError)

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

	confUpload := model.GetConfUpload(*inputDir)

	_, err = subcommand.UploadCarFiles(confUpload)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}

//python3 swan_cli.py task --input-dir /home/peware/testGoSwanProvider/output --out-dir /home/peware/testGoSwanProvider/task --miner t03354 --dataset test --description test
func createTask() error {
	cmd := flag.NewFlagSet(subcommand.SUBCOMMAND_TASK, flag.ExitOnError)

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

	confTask := model.GetConfTask(*inputDir, outputDir, *taskName, *minerFid, *dataset, *description)
	confDeal := model.GetConfDeal(outputDir, *minerFid, "")
	jsonFileName, _, _, err := subcommand.CreateTask(confTask, confDeal)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	logs.GetLogger().Info("Task information is in:", *jsonFileName)

	return nil
}

func sendDeal() error {
	cmd := flag.NewFlagSet(subcommand.SUBCOMMAND_DEAL, flag.ExitOnError)

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

	_, err = subcommand.SendDealsByConfig(*outputDir, *minerFid, *metadataJsonPath)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}

func sendAutoBidDeal() error {
	cmd := flag.NewFlagSet(subcommand.SUBCOMMAND_DEAL, flag.ExitOnError)

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

	subcommand.SendAutoBidDealsLoopByConfig(*outputDir)
	return nil
}
