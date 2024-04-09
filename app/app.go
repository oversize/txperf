package app

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/blinklabs-io/bursa"
)

type TxPerfApp struct {
	Name   string
	Config *Config
	Wallet *bursa.Wallet
}

var app = TxPerfApp{
	Name: "Bar",
}

func GetApp() *TxPerfApp {
	return &app
}

func LoadApp() error {
	err := LoadConfig()
	if err != nil {
		return err
	}
	return nil
}

// CreateNewKey creates a new key with given name and stores it in ~/.txperf/wallets/wallet.json
func (a TxPerfApp) CreateNewWallet(walletName string) error {

	wallet, err := a.newBursaWallet(walletName)
	if err != nil {
		return fmt.Errorf("unable to create new bursa wallet: %s", err.Error())
	}

	fmt.Println(wallet)

	// Call bursa to create the keys
	// write key material to the files
	return nil
}

// createWallet creates a wallet by using bursa to generate a mnemnic and
// accompanying key material.
func (a TxPerfApp) newBursaWallet(walletName string) (*bursa.Wallet, error) {

	walletFile := GetConfig().WalletDir + "/" + walletName + ".json"
	log.Default().Println(walletFile)

	if _, err := os.Stat(walletFile); err == nil {
		return nil, fmt.Errorf("wallet already exists")
	}

	mnemonic, err := bursa.NewMnemonic()
	if err != nil {
		return nil, fmt.Errorf("failed loading config from environment: %s", err)
	}

	w, err := bursa.NewWallet(mnemonic, GetConfig().Network, 0, 0, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("unable to create wallet: %s", err)
	}

	f, _ := json.MarshalIndent(*w, "", "  ")

	err = os.WriteFile(walletFile, f, 0600)
	if err != nil {
		return nil, fmt.Errorf("unable to write wallet.json: %s", err.Error())
	}

	return w, nil
}

func (a TxPerfApp) Run() error {
	walletPath := GetConfig().GetWalletPath("mywallet1")
	wf, err := os.ReadFile(walletPath)
	if err != nil {
		return fmt.Errorf("unable to read wallet file %s", err.Error())
	}
	nw := bursa.Wallet{}
	json.Unmarshal(wf, &nw)

	fmt.Println(nw)
	return nil
}
