package app

import (
	"fmt"
	"log"
	"os"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Network          string
	Magic            int
	RootDir          string
	WalletDir        string
	BlockfrostApiKey string `envconfig:"BLOCKFROST_API_KEY"`
}

var globalConfig = &Config{
	Network: "preprod",
	Magic:   1,
	//BlockfrostApiKey: "preprodzu4liHARfKdaVvi27p1T0sFarzZlDaZP",
}

func GetConfig() *Config {
	return globalConfig
}

func LoadConfig() error {

	err := envconfig.Process("txperf", globalConfig)
	if err != nil {
		return fmt.Errorf("error processing environment: %s", err)
	}

	rootDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("unable to get user home dir %v", err.Error())
	}
	rootDir += "/.txperf"
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		os.Mkdir(rootDir, 0700)
	}

	walletDir := rootDir + "/wallets"
	if _, err := os.Stat(walletDir); os.IsNotExist(err) {
		os.Mkdir(walletDir, 0700)
	}

	globalConfig.RootDir = rootDir
	globalConfig.WalletDir = walletDir

	return nil
}

func (c Config) GetWalletPath(walletName string) string {
	walletPath := GetConfig().WalletDir + "/" + walletName + ".json"
	if _, err := os.Stat(walletPath); os.IsNotExist(err) {
		log.Fatalf("wallet does not exist: %s", err.Error())
	}
	return walletPath
}
