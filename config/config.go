package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/filswan/go-swan-lib/logs"
)

type Configuration struct {
	Lotus      lotus      `toml:"lotus"`
	Main       main       `toml:"main"`
	IpfsServer ipfsServer `toml:"ipfs_server"`
	Sender     sender     `toml:"sender"`
}

type lotus struct {
	ClientApiUrl      string `toml:"client_api_url"`
	ClientAccessToken string `toml:"client_access_token"`
}

type main struct {
	SwanApiUrl      string `toml:"api_url"`
	SwanApiKey      string `toml:"api_key"`
	SwanAccessToken string `toml:"access_token"`
	SwanRepo        string `toml:"swan_repo"`
	MarketVersion   string `toml:"market_version"`
}

type ipfsServer struct {
	DownloadUrlPrefix string `toml:"download_url_prefix"`
	UploadUrlPrefix   string `toml:"upload_url_prefix"`
}

type sender struct {
	OfflineSwan           bool          `toml:"offline_swan"`
	VerifiedDeal          bool          `toml:"verified_deal"`
	FastRetrieval         bool          `toml:"fast_retrieval"`
	SkipConfirmation      bool          `toml:"skip_confirmation"`
	GenerateMd5           bool          `toml:"generate_md5"`
	Wallet                string        `toml:"wallet"`
	MaxPrice              string        `toml:"max_price"`
	StartEpochHours       int           `toml:"start_epoch_hours"`
	ExpireDays            int           `toml:"expire_days"`
	Duration              int           `toml:"duration"`
	StartDealTimeInterval time.Duration `toml:"start_deal_time_interval"`
}

type ChainInfo struct {
	ChainRpcServices []ChainRpcService `json:"chain_rpc_services"`
}
type ChainRpcService struct {
	Remark      string   `json:"remark"`
	ChainName   string   `json:"chain_name"`
	RpcEndpoint []string `json:"rpc_endpoint"`
}

var config *Configuration
var chainInfo *ChainInfo

func initConfig() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		logs.GetLogger().Fatal("Cannot get home directory.")
	}

	configFile := filepath.Join(homedir, ".swan/client/config.toml")
	if metaData, err := toml.DecodeFile(configFile, &config); err != nil {
		log.Fatal("Error:", err)
	} else {
		if !requiredFieldsAreGiven(metaData) {
			log.Fatal("Required fields not given")
		}
	}
	config.Main.SwanRepo = filepath.Join(homedir, ".swan/client/boost")
}

func initChain() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		logs.GetLogger().Fatal("Cannot get home directory.")
	}

	chainInfoPath := filepath.Join(homedir, ".swan/client/chain-rpc.json")
	chainFile, err := os.Open(chainInfoPath)
	if err != nil {
		log.Fatalf("open chain-rpc.json failed,error: %v", err)
	}
	chainBytes, err := io.ReadAll(chainFile)
	if err != nil {
		log.Fatalf("read chain-rpc.json failed,error: %v", err)
	}
	if err := json.Unmarshal(chainBytes, &chainInfo); err != nil {
		log.Fatal("Error:", err)
	}
}

func GetChainByChainName(name string) (chain ChainRpcService, err error) {
	chains := GetChain()
	for _, c := range chains.ChainRpcServices {
		if strings.EqualFold(c.ChainName, name) {
			chain = c
			return chain, nil
		}
	}
	return ChainRpcService{}, errors.New(fmt.Sprintf("not support chainName: %s", chain))
}

func GetChain() ChainInfo {
	if chainInfo == nil {
		initChain()
	}
	return *chainInfo
}

func GetConfig() Configuration {
	if config == nil {
		initConfig()
	}
	return *config
}

func requiredFieldsAreGiven(metaData toml.MetaData) bool {
	requiredFields := [][]string{
		{"lotus"},
		{"main"},
		{"ipfs_server"},
		{"sender"},

		{"lotus", "client_api_url"},
		{"lotus", "client_access_token"},

		{"main", "api_url"},
		{"main", "api_key"},
		{"main", "access_token"},
		{"main", "market_version"},

		{"ipfs_server", "download_url_prefix"},
		{"ipfs_server", "upload_url_prefix"},

		{"sender", "offline_swan"},
		{"sender", "verified_deal"},
		{"sender", "fast_retrieval"},
		{"sender", "skip_confirmation"},
		{"sender", "generate_md5"},
		{"sender", "wallet"},
		{"sender", "max_price"},
		{"sender", "start_epoch_hours"},
		{"sender", "expire_days"},
		{"sender", "duration"},
		{"sender", "start_deal_time_interval"},
	}

	for _, v := range requiredFields {
		if !metaData.IsDefined(v...) {
			log.Fatal("Required fields ", v)
		}
	}

	return true
}
