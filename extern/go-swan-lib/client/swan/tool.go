package swan

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/filswan/go-swan-lib/client/web"
	"github.com/filswan/go-swan-lib/constants"
	"github.com/filswan/go-swan-lib/logs"
	"github.com/filswan/go-swan-lib/utils"
)

func (swanClient *SwanClient) CheckDatacap(wallet string) (bool, error) {
	apiUrl := swanClient.ApiUrl + "/tools/check_datacap?address=" + wallet
	params := url.Values{}

	response, err := web.HttpGetNoToken(apiUrl, strings.NewReader(params.Encode()))

	if err != nil {
		logs.GetLogger().Error(err)
		return false, err
	}

	status := utils.GetFieldStrFromJson(response, "status")

	if !strings.EqualFold(status, constants.SWAN_API_STATUS_SUCCESS) {
		message := utils.GetFieldStrFromJson(response, "message")
		err := fmt.Errorf("error:%s,%s", status, message)
		logs.GetLogger().Error(err)
		return false, err
	}

	data := utils.GetFieldMapFromJson(response, "data")
	isVerified := data["is_verified"].(bool)

	return isVerified, nil
}

func (swanClient *SwanClient) StatisticsChainInfo(chainId string) error {
	chainName, ok := constants.ChainMap[chainId]
	if !ok {
		return errors.New(fmt.Sprintf("not support chainId: %s", chainId))
	}
	err := swanClient.GetJwtTokenUp3Times()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	var req struct {
		ChainName string `json:"chain_name"`
		UserKey   string `json:"user_key"`
	}
	req.UserKey = swanClient.ApiKey
	req.ChainName = chainName
	reqParam, _ := json.Marshal(req)
	buffer := bytes.NewBuffer(reqParam)

	client := http.Client{
		Timeout: 30 * time.Second,
	}
	apiUrl := utils.UrlJoin(swanClient.ApiUrl, "statistics/chain")
	if _, err := client.Post(apiUrl, "application/json", buffer); err != nil {
		return err
	}
	return nil
}

func (swanClient *SwanClient) StatisticsNodeStatus() error {
	err := swanClient.GetJwtTokenUp3Times()
	if err != nil {
		logs.GetLogger().Error(err)
		return err
	}

	var req struct {
		UserKey string `json:"user_key"`
	}
	req.UserKey = swanClient.ApiKey
	reqParam, _ := json.Marshal(req)
	buffer := bytes.NewBuffer(reqParam)

	client := http.Client{
		Timeout: 30 * time.Second,
	}
	apiUrl := utils.UrlJoin(swanClient.ApiUrl, "statistics/node")
	if _, err = client.Post(apiUrl, "application/json", buffer); err != nil {
		return err
	}
	return nil
}
