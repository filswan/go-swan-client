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

	swanClient, err := swan.GetClient(confDeal.SwanApiUrl, confDeal.SwanApiKey, confDeal.SwanAccessToken, confDeal.SwanToken)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	sourceId := libconstants.TASK_SOURCE_ID_SWAN_CLIENT
	dealStatus := libconstants.OFFLINE_DEAL_STATUS_ASSIGNED
	assignedOfflineDeals, err := swanClient.GetOfflineDealsByStatus(dealStatus, nil, &sourceId, nil, nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	if len(assignedOfflineDeals) == 0 {
		logs.GetLogger().Info("no offline deals to be sent")
		return nil, nil
	}

	var tasksDeals [][]*libmodel.FileDesc
	for _, assignedOfflineDeal := range assignedOfflineDeals {
		updateOfflineDealParams, err := SendAutobidDeal(confDeal, assignedOfflineDeal)
		if err != nil {
			logs.GetLogger().Error(err)
			continue
		}

		err = swanClient.UpdateOfflineDeal(*updateOfflineDealParams)
		if err != nil {
			logs.GetLogger().Error(err)
			continue
		}
	}

	return tasksDeals, nil
}

func SendAutobidDeal(confDeal *model.ConfDeal, offlineDeal *libmodel.OfflineDeal) (*swan.UpdateOfflineDealParams, error) {
	if confDeal == nil {
		err := fmt.Errorf("parameter confDeal is nil")
		logs.GetLogger().Error(err)
		return nil, err
	}

	offlineDeal.DealCid = strings.Trim(offlineDeal.DealCid, " ")
	if len(offlineDeal.DealCid) != 0 {
		return nil, nil
	}

	err := model.SetDealConfig4Autobid(confDeal, *offlineDeal)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	for i := 0; i < 60; i++ {
		msg := fmt.Sprintf("send deal for task:%s, deal:%d", *offlineDeal.TaskUuid, offlineDeal.Id)
		logs.GetLogger().Info(msg)
		dealConfig := libmodel.DealConfig{
			VerifiedDeal:     confDeal.VerifiedDeal,
			FastRetrieval:    confDeal.FastRetrieval,
			SkipConfirmation: confDeal.SkipConfirmation,
			MinerPrice:       confDeal.MinerPrice,
			StartEpoch:       int(confDeal.StartEpoch),
			MinerFid:         confDeal.MinerFid,
			SenderWallet:     confDeal.SenderWallet,
			Duration:         int(confDeal.Duration),
			TransferType:     libconstants.LOTUS_TRANSFER_TYPE_MANUAL,
			PayloadCid:       offlineDeal.PayloadCid,
			PieceCid:         offlineDeal.PieceCid,
			FileSize:         offlineDeal.CarFileSize,
		}

		err = CheckDealConfig(confDeal, &dealConfig)
		if err != nil {
			logs.GetLogger().Error(err)
			continue
		}

		lotusClient, err := lotus.LotusGetClient(confDeal.LotusClientApiUrl, confDeal.LotusClientAccessToken)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}

		dealCid, startEpoch, err := lotusClient.LotusClientStartDeal(dealConfig, i)
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

		updateOfflineDealParams := swan.UpdateOfflineDealParams{
			DealId:     offlineDeal.Id,
			DealCid:    dealCid,
			Status:     libconstants.OFFLINE_DEAL_STATUS_CREATED,
			StartEpoch: startEpoch,
		}

		logs.GetLogger().Info("task:", offlineDeal.TaskUuid, ", deal CID:", dealInfo.DealCid, ", start epoch:", dealInfo.StartEpoch, ", deal sent to ", confDeal.MinerFid, " successfully")

		return &updateOfflineDealParams, nil
	}

	err = fmt.Errorf("failed to send deal")
	logs.GetLogger().Error(err)
	return nil, err
}
