package subcommand

import (
	"fmt"
	"strings"
	"time"

	"github.com/filswan/go-swan-client/model"

	"github.com/filswan/go-swan-lib/logs"

	"github.com/filswan/go-swan-lib/client/lotus"
	"github.com/filswan/go-swan-lib/client/swan"
	libconstants "github.com/filswan/go-swan-lib/constants"
	libmodel "github.com/filswan/go-swan-lib/model"
	"github.com/filswan/go-swan-lib/utils"
)

func SendAutoBidDealsLoopByConfig(outputDir string) error {
	confDeal := model.GetConfDeal(&outputDir, "", "")
	err := SendAutoBidDealsLoop(confDeal)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	return nil
}

func SendAutoBidDealsLoop(confDeal *model.ConfDeal) error {
	err := CreateOutputDir(confDeal.OutputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	for {
		_, err := SendAutoBidDeals(confDeal)
		if err != nil {
			logs.GetLogger().Error(err)
			//return err
			continue
		}

		time.Sleep(time.Second * 30)
	}
}

func SendAutoBidDeals(confDeal *model.ConfDeal) ([][]*libmodel.FileDesc, error) {
	if confDeal == nil {
		err := fmt.Errorf("parameter confDeal is nil")
		logs.GetLogger().Error(err)
		return nil, err
	}

	err := CreateOutputDir(confDeal.OutputDir)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	logs.GetLogger().Info("output dir is:", confDeal.OutputDir)

	swanClient, err := swan.GetClient(confDeal.SwanApiUrlToken, confDeal.SwanApiUrl, confDeal.SwanApiKey, confDeal.SwanAccessToken, confDeal.SwanToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	params := swan.GetOfflineDealsByStatusParams{
		DealStatus: libconstants.OFFLINE_DEAL_STATUS_ASSIGNED,
	}
	assignedOfflineDeals, err := swanClient.GetOfflineDealsByStatus(params)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if len(assignedOfflineDeals) == 0 {
		logs.GetLogger().Info("no autobid offline deals to be sent")
		return nil, nil
	}

	var tasksDeals [][]*libmodel.FileDesc
	for _, assignedOfflineDeal := range assignedOfflineDeals {
		_, _, err := SendAutobidDeal(confDeal, &assignedOfflineDeal)
		if err != nil {
			logs.GetLogger().Error(err)
			continue
		}
	}

	return tasksDeals, nil
}

func SendAutobidDeal(confDeal *model.ConfDeal, offlineDeal *libmodel.OfflineDeal) (int, *libmodel.FileDesc, error) {
	if confDeal == nil {
		err := fmt.Errorf("parameter confDeal is nil")
		logs.GetLogger().Error(err)
		return 0, nil, err
	}

	fileDesc := libmodel.FileDesc{
		Deals: []*libmodel.DealInfo{},
	}
	dealSentNum := 0
	offlineDeal.DealCid = strings.Trim(offlineDeal.DealCid, " ")
	if len(offlineDeal.DealCid) != 0 {
		return 0, nil, nil
	}

	err := model.SetDealConfig4Autobid(confDeal, *offlineDeal)
	if err != nil {
		logs.GetLogger().Error(err)
		return 0, nil, nil
	}

	fileSizeInt := utils.GetInt64FromStr(offlineDeal.FileSize)
	if fileSizeInt <= 0 {
		logs.GetLogger().Error("file is too small")
		return 0, nil, nil
	}
	pieceSize, sectorSize := utils.CalculatePieceSize(fileSizeInt)
	cost := utils.CalculateRealCost(sectorSize, confDeal.MinerPrice)
	for i := 0; i < 60; i++ {
		msg := fmt.Sprintf("send deal for task:%s, deal:%d", *offlineDeal.TaskUuid, offlineDeal.Id)
		logs.GetLogger().Info(msg)
		dealConfig := libmodel.GetDealConfig(confDeal.VerifiedDeal, confDeal.FastRetrieval, confDeal.SkipConfirmation, confDeal.MinerPrice, int(confDeal.StartEpoch), int(confDeal.Duration), confDeal.MinerFid, confDeal.SenderWallet)

		err = CheckDealConfig(confDeal, dealConfig)
		if err != nil {
			logs.GetLogger().Error(err)
			continue
		}

		lotusClient, err := lotus.LotusGetClient(confDeal.LotusClientApiUrl, confDeal.LotusClientAccessToken)
		if err != nil {
			logs.GetLogger().Error(err)
			return 0, nil, err
		}

		dealCid, startEpoch, err := lotusClient.LotusClientStartDeal(fileDesc, cost, pieceSize, *dealConfig, i)
		if err != nil {
			logs.GetLogger().Error("tried ", i, " times,", err)

			if strings.Contains(err.Error(), "already tracking identifier") {
				continue
			} else {
				break
			}
		}
		if dealCid == nil {
			logs.GetLogger().Info("no deal CID returned")
			continue
		}

		dealInfo := &libmodel.DealInfo{
			MinerFid:   offlineDeal.MinerFid,
			DealCid:    *dealCid,
			StartEpoch: *startEpoch,
		}
		fileDesc.Deals = append(fileDesc.Deals, dealInfo)
		dealSentNum = dealSentNum + 1

		logs.GetLogger().Info("task:", offlineDeal.TaskUuid, ", deal CID:", dealInfo.DealCid, ", start epoch:", dealInfo.StartEpoch, ", deal sent to ", confDeal.MinerFid, " successfully")
		break
	}
	return dealSentNum, &fileDesc, nil
}
