package main

import (
	"github.com/filswan/go-swan-lib/client/web"
	"github.com/filswan/go-swan-lib/logs"
)

func main() {
	response, err := web.HttpGetNoToken("https://calibration-api.filscout.com/api/v1/storagedeal/6666", nil)
	if err != nil {
		logs.GetLogger().Error(err)
		return
	}

	logs.GetLogger().Info(response)
}
