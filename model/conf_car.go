package model

import (
	"path/filepath"
	"time"

	"github.com/DoraNebula/go-swan-client/config"
)

type ConfCar struct {
	LotusApiUrl      string
	LotusAccessToken string
	OutputDir        string
	InputDir         string
}

func GetConfCar(inputDir string, outputDir *string) *ConfCar {
	confCar := &ConfCar{
		LotusApiUrl:      config.GetConfig().Lotus.ApiUrl,
		LotusAccessToken: config.GetConfig().Lotus.AccessToken,
		OutputDir:        filepath.Join(config.GetConfig().Sender.OutputDir, time.Now().Format("2006-01-02_15:04:05")),
		InputDir:         inputDir,
	}

	if outputDir != nil && len(*outputDir) != 0 {
		confCar.OutputDir = *outputDir
	}

	return confCar
}
