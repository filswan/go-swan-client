package model

import (
	"go-swan-client/config"
	"time"
)

type ConfCar struct {
	LotusApiUrl      string
	LotusAccessToken string
	OutputDir        string
}

func GetConfCar(outDir *string) *ConfCar {
	confCar := &ConfCar{
		LotusApiUrl:      config.GetConfig().Lotus.ApiUrl,
		LotusAccessToken: config.GetConfig().Lotus.AccessToken,
		OutputDir:        config.GetConfig().Sender.OutputDir + time.Now().Format("2006-01-02_15:04:05"),
	}

	if outDir != nil {
		confCar.OutputDir = *outDir
	}

	return confCar
}
