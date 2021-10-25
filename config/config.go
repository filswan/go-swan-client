package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/DoraNebula/go-swan-client/logs"

	"github.com/BurntSushi/toml"
)

type Configuration struct {
	Lotus      lotus      `toml:"lotus"`
	Main       main       `toml:"main"`
	WebServer  webServer  `toml:"web_server"`
	IpfsServer ipfsServer `toml:"ipfs_server"`
	Sender     sender     `toml:"sender"`
}

type lotus struct {
	ApiUrl      string `toml:"api_url"`
	AccessToken string `toml:"access_token"`
}

type main struct {
	SwanApiUrl        string `toml:"api_url"`
	SwanApiKey        string `toml:"api_key"`
	SwanAccessToken   string `toml:"access_token"`
	StorageServerType string `toml:"storage_server_type"`
}

type webServer struct {
	DownloadUrlPrefix string `toml:"download_url_prefix"`
}

type ipfsServer struct {
	DownloadUrlPrefix string `toml:"download_url_prefix"`
}

type sender struct {
	BidMode            int    `toml:"bid_mode"`
	OfflineMode        bool   `toml:"offline_mode"`
	OutputDir          string `toml:"output_dir"`
	PublicDeal         bool   `toml:"public_deal"`
	VerifiedDeal       bool   `toml:"verified_deal"`
	FastRetrieval      bool   `toml:"fast_retrieval"`
	SkipConfirmation   bool   `toml:"skip_confirmation"`
	GenerateMd5        bool   `toml:"generate_md5"`
	Wallet             string `toml:"wallet"`
	MaxPrice           string `toml:"max_price"`
	StartEpochHours    int    `toml:"start_epoch_hours"`
	ExpireDays         int    `toml:"expire_days"`
	GocarFileSizeLimit int64  `toml:"gocar_file_size_limit"`
}

var config *Configuration

func initConfig() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		logs.GetLogger().Fatal("Cannot get home directory.")
	}

	configFile := filepath.Join(homedir, ".swan/client/config.toml")
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
		initConfig()
	}
	return *config
}

func requiredFieldsAreGiven(metaData toml.MetaData) bool {
	requiredFields := [][]string{
		{"lotus"},
		{"main"},
		{"web_server"},
		{"ipfs_server"},
		{"sender"},

		{"lotus", "api_url"},
		{"lotus", "access_token"},
		{"lotus", "miner_api_url"},

		{"main", "api_url"},
		{"main", "api_key"},
		{"main", "access_token"},
		{"main", "storage_server_type"},

		{"web_server", "download_url_prefix"},

		{"ipfs_server", "download_url_prefix"},

		{"sender", "bid_mode"},
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
		{"sender", "expire_days"},
		{"sender", "gocar_file_size_limit"},
	}

	for _, v := range requiredFields {
		if !metaData.IsDefined(v...) {
			log.Fatal("required fields ", v)
		}
	}

	return true
}
