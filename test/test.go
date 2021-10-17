package test

import (
	"fmt"
	"path/filepath"
	"strings"

	"go-swan-client/common/client"
	"go-swan-client/logs"
	"go-swan-client/model"
	"go-swan-client/subcommand"
)

func Test() {
	TestLotusClient()
}

func TestLotusClient() {
	result := client.LotusMarketGetAsk()
	logs.GetLogger().Info(*result)
	result1 := client.LotusClientCalcCommP("/home/peware/go-swan-client/carFiles/hello2.txt.car")
	logs.GetLogger().Info(*result1)
	result2 := client.LotusClientImport("/home/peware/go-swan-client/carFiles/hello2.txt.car", true)
	logs.GetLogger().Info(*result2)
	client.LotusClientGenCar("/home/peware/go-swan-client/srcFiles/hello2.txtd", "/home/peware/go-swan-client/srcFiles/hello2.txt.car", false)
	version, err := client.LotusVersion()
	logs.GetLogger().Info(*version, err)
	cid, err := client.LotusClientStartDeal("t03354", "bafykbzaceciszub37cz6vtj2vq2x3cofgiiksqsmym3k23cf2jyhnq5dwvvas", "baga6ea4seaqh2pi3qfhhghuxuz2iwpclr6xfosdzo5nd2sdjqynh3ddvkrorgla", "t3u7pumush376xbytsgs5wabkhtadjzfydxxda2vzyasg7cimkcphswrq66j4dubbhwpnojqd3jie6ermpwvvq", "0", 1024, 10101, true, true)
	if cid != nil {
		logs.GetLogger().Info(*cid)
	}
	if err != nil {
		logs.GetLogger().Error(err)
	}

}

func TestGetTasks() {
	swanClient := client.SwanGetClient()
	swanClient.GetAssignedTasks()
}

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
		IsPublic:       1,
		//IsVerified:     true,
		MinerFid: &minerId,
	}

	swan := client.SwanGetClient()

	response := swan.SwanCreateTask(task, "/Users/dorachen/go-workspace/src/go-swan-client/test/car.csv")
	logs.GetLogger().Info(response)
}

func TestFilePath() {
	filename := filepath.Base("/Users/dorachen/go-workspace/src/go-swan-client/test.go")
	logs.GetLogger().Info(filename)
	logs.GetLogger().Info(strings.TrimSuffix(filename, filepath.Ext(filename)))
	logs.GetLogger().Info(filepath.Join("/abc////", "path2"))
	logs.GetLogger().Info(filepath.Join("/abc////", ""))
	logs.GetLogger().Info(filepath.Join("/abc////", "", "test"))
}

func TestDealConfig() {
	dealConfig := subcommand.GetDealConfig("t03354")
	subcommand.CheckDealConfig(dealConfig)
	pieceSize, sectorSize := subcommand.CalculatePieceSize(2049)
	cost := subcommand.CalculateRealCost(sectorSize, dealConfig.MinerPrice)
	msg := fmt.Sprintf("Piece size:%d, sector size:%f,cost:%f", pieceSize, sectorSize, cost)
	logs.GetLogger().Info(msg)
}
