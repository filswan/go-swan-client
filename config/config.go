package config

import (
	"log"
	"strings"

	"github.com/BurntSushi/toml"
)

type Configuration struct {
	Main       main       `toml:"main"`
	WebServer  webServer  `toml:"web_server"`
	IpfsServer ipfsServer `toml:"ipfs_server"`
	Sender     sender     `toml:"sender"`
}

type main struct {
	SwanApiUrl        string `toml:"api_url"`
	SwanApiKey        string `toml:"api_key"`
	SwanAccessToken   string `toml:"access_token"`
	StorageServerType string `toml:"storage_server_type"`
}

type webServer struct {
	Host string `toml:"host"`
	Port int    `toml:"port"`
	Path string `toml:"path"`
}

type ipfsServer struct {
	GatewayAddress string `toml:"gateway_address"`
}

type sender struct {
	OfflineMode      bool   `toml:"offline_mode"`
	OutputDir        string `toml:"output_dir"`
	PublicDeal       bool   `toml:"public_deal"`
	VerifiedDeal     bool   `toml:"verified_deal"`
	FastRetrieval    bool   `toml:"fast_retrieval"`
	SkipConfirmation bool   `toml:"skip_confirmation"`
	GenerateMd5      bool   `toml:"generate_md5"`
	Wallet           string `toml:"wallet"`
	MaxPrice         string `toml:"max_price"`
	StartEpochHours  int    `toml:"start_epoch_hours"`
}

var config *Configuration

func InitConfig(configFile string) {
	if strings.Trim(configFile, " ") == "" {
		configFile = "./config/config.toml"
	}
	if metaData, err := toml.DecodeFile(configFile, &config); err != nil {
		log.Fatal("error:", err)
	} else {
		if !requiredFieldsAreGiven(metaData) {
			log.Fatal("required fields not given")
		}
	}
}

func GetConfig() Configuration {
	if config == nil {
		InitConfig("")
	}
	return *config
}

func requiredFieldsAreGiven(metaData toml.MetaData) bool {
	requiredFields := [][]string{
		{"main"},
		{"web_server"},
		{"ipfs_server"},
		{"sender"},

		{"main", "api_url"},
		{"main", "api_key"},
		{"main", "access_token"},
		{"main", "storage_server_type"},

		{"web_server", "host"},
		{"web_server", "port"},
		{"web_server", "path"},

		{"ipfs_server", "gateway_address"},

		{"sender", "offline_mode"},
		{"sender", "output_dir"},
		{"sender", "public_deal"},
		{"sender", "verified_deal"},
		{"sender", "fast_retrieval"},
		{"sender", "skip_confirmation"},
		{"sender", "generate_md5"},
		{"sender", "wallet"},
		{"sender", "max_price"},
		{"sender", "start_epoch_hours"},
	}

	for _, v := range requiredFields {
		if !metaData.IsDefined(v...) {
			log.Fatal("required fields ", v)
		}
	}

	return true
}
