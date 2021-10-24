package model

import "go-swan-client/config"

type ConfCar struct {
	LotusApiUrl      string
	LotusAccessToken string
}

func GetConfCar() *ConfCar {
	confCar := &ConfCar{
		LotusApiUrl:      config.GetConfig().Lotus.ApiUrl,
		LotusAccessToken: config.GetConfig().Lotus.AccessToken,
	}

	return confCar
}
