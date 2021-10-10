package subcommand

import (
	"encoding/csv"
	"go-swan-client/common/utils"
	"go-swan-client/config"
	"go-swan-client/logs"
	"go-swan-client/model"
	"os"
)

func sendDeals(outputDir *string, task model.Task, carFiles []*model.FileDesc, taskUuid string) {
	//fromWallet := config.GetConfig().Sender.Wallet
	//maxPrice := config.GetConfig().Sender.MaxPrice
	//verifiedDeal := config.GetConfig().Sender.VerifiedDeal
	//fastRetrieval := config.GetConfig().Sender.FastRetrieval
	//epochIntervalHours := config.GetConfig().Sender.StartEpochHours

	if outputDir == nil {
		outDir := config.GetConfig().Sender.OutputDir
		outputDir = &outDir
	}

	//deal_config = DealConfig(miner_id, fromWallet, maxPrice, verifiedDeal, fastRetrieval, epochIntervalHours)

	sendDeals2Miner(task, *outputDir, carFiles, taskUuid)
}

func sendDeals2Miner(task model.Task, outputDir string, carFiles []*model.FileDesc, taskUuid string) {
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}

	//skipConfirmation := config.GetConfig().Sender.SkipConfirmation

	minerId := ""

	err = createCsv4SendDeal(carFiles, &minerId, outputDir, nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}

}

func createCsv4SendDeal(carFiles []*model.FileDesc, minerId *string, outDir string, task *model.Task) error {
	csvFileName := task.TaskName + ".csv"
	csvFilePath := utils.GetPath(outDir, csvFileName)

	logs.GetLogger().Info("Swan task CSV Generated: ", csvFilePath)

	headers := []string{
		"uuid",
		"miner_id",
		"file_source_url",
		"md5",
		"start_epoch",
		"deal_cid",
	}

	file, err := os.Create(csvFilePath)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.Write(headers)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	for _, carFile := range carFiles {
		columns := []string{}
		columns = append(columns, carFile.Uuid)
		if minerId != nil {
			columns = append(columns, *minerId)
		} else {
			columns = append(columns, "")
		}
		columns = append(columns, carFile.CarFileUrl)
		columns = append(columns, carFile.CarFileMd5)
		columns = append(columns, carFile.StartEpoch)
		columns = append(columns, carFile.DealCid)

		err = writer.Write(columns)
		if err != nil {
			logs.GetLogger().Error(err)
			return err
		}
	}

	return nil
}
