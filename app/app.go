package app

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/Salvionied/apollo"
	"github.com/Salvionied/apollo/constants"
	serAddress "github.com/Salvionied/apollo/serialization/Address"
	"github.com/Salvionied/apollo/serialization/Key"
	"github.com/Salvionied/apollo/serialization/UTxO"
	"github.com/Salvionied/apollo/txBuilding/Backend/BlockFrostChainContext"
	"github.com/blinklabs-io/bursa"
	ouroboros "github.com/blinklabs-io/gouroboros"
	"github.com/blinklabs-io/gouroboros/ledger"
	"github.com/blinklabs-io/gouroboros/protocol/txsubmission"
)

type TxPerfApp struct {
	Name   string
	Config *Config
	Wallet *bursa.Wallet
}

var app = TxPerfApp{
	Name: "Bar",
}
var ntnTxBytes []byte
var ntnTxHash [32]byte
var ntnTxType uint
var ntnSentTx bool
var ntnDoneChan chan any

func GetApp() *TxPerfApp {
	return &app
}

func LoadApp() error {
	err := LoadConfig()
	if err != nil {
		return fmt.Errorf("could not load config: %s", err.Error())
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

// createWallet creates a wallet by using bursa to generate a mnemonic and
// stores it unecrypted in ~/.txperf/wallets/[walletName].json
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

// Run runs txperf.
func (a TxPerfApp) Run() error {
	w1, w2, err := loadWallets()
	if err != nil {
		return fmt.Errorf("unable to load wallets %s", err.Error())
	}

	apolloEmptyBackend := apollo.NewEmptyBackend()
	apolloBackend := apollo.New(&apolloEmptyBackend)
	apolloBackend = apolloBackend.
		SetWalletFromBech32(w1.PaymentAddress).
		SetWalletAsChangeAddress()

	utxos, err := getUtxosByAddress(w1.PaymentAddress)
	if err != nil {
		return fmt.Errorf("could not get utxos: %s", err.Error())
	}
	fmt.Println(utxos)

	apolloBackend = apolloBackend.AddLoadedUTxOs(utxos...)
	apolloBackend = apolloBackend.PayToAddressBech32(w2.PaymentAddress, int(20000000))
	apolloTx, err := apolloBackend.Complete()
	if err != nil {
		return fmt.Errorf("could not complete tx: %s", err.Error())
	}

	vKeyBytes, err := hex.DecodeString(w1.PaymentVKey.CborHex)
	if err != nil {
		return fmt.Errorf("could not decode payment vkey: %s", err.Error())
	}
	sKeyBytes, err := hex.DecodeString(w1.PaymentExtendedSKey.CborHex)
	if err != nil {
		return fmt.Errorf("could not decode payment extended skey: %s", err.Error())
	}

	vKeyBytes = vKeyBytes[2:]
	sKeyBytes = sKeyBytes[2:]
	sKeyBytes = append(sKeyBytes[:64], sKeyBytes[96:]...)
	vkey := Key.VerificationKey{Payload: vKeyBytes}
	skey := Key.SigningKey{Payload: sKeyBytes}
	apolloTx, err = apolloTx.SignWithSkey(vkey, skey)
	if err != nil {
		return fmt.Errorf("could not sign with skey: %s", err.Error())
	}

	// Gets *Transaction.Transaction for sending to gorouboros
	tx := apolloTx.GetTx()
	txBytes, err := tx.Bytes()
	if err != nil {
		return err
	}
	if err := SubmitTx(txBytes); err != nil {
		return err
	}

	return nil
}

func SubmitTx(txBytes []byte) error {
	fmt.Println("Submitting tx")
	cfg := GetConfig()
	ntnTxBytes = txBytes[:]
	ntnSentTx = false

	txType, err := ledger.DetermineTransactionType(txBytes)
	if err != nil {
		return fmt.Errorf("could not parse transaction to determine type: %s", err)
	}

	tx, err := ledger.NewTransactionFromCbor(txType, txBytes)
	if err != nil {
		return fmt.Errorf("failed to parse transaction CBOR: %s", err)
	}
	txHashHex, err := hex.DecodeString(tx.Hash())
	if err != nil {
		return fmt.Errorf("failed to decode TX hash: %s", err)
	}
	ntnTxHash = [32]byte(txHashHex)
	ntnTxType = txType

	conn, err := createClientConnection("127.0.0.1:3001")
	if err != nil {
		return err
	}
	errorChan := make(chan error)
	// Capture errors
	go func() {
		err, ok := <-errorChan
		if ok {
			panic(fmt.Errorf("async: %s", err))
		}
	}()
	oConn, err := ouroboros.New(
		ouroboros.WithConnection(conn),
		ouroboros.WithNetwork(
			ouroboros.NetworkByName(cfg.Network),
		),
		ouroboros.WithErrorChan(errorChan),
		ouroboros.WithNodeToNode(true),
		ouroboros.WithKeepAlive(true),
		ouroboros.WithTxSubmissionConfig(
			txsubmission.NewConfig(
				txsubmission.WithRequestTxIdsFunc(handleRequestTxIds),
				txsubmission.WithRequestTxsFunc(handleRequestTxs),
			),
		),
	)
	if err != nil {
		return err
	}

	// Start txSubmission loop
	ntnDoneChan = make(chan any)
	oConn.TxSubmission().Client.Init()
	<-ntnDoneChan
	// Sleep 2s to allow time for TX to enter remote mempool before closing connection
	time.Sleep(2 * time.Second)

	if err := oConn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %s", err)
	}

	return nil
}

func createClientConnection(nodeAddress string) (net.Conn, error) {
	var err error
	var conn net.Conn
	var dialProto string
	var dialAddress string
	dialProto = "tcp"
	dialAddress = nodeAddress
	conn, err = net.Dial(dialProto, dialAddress)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func loadWallets() (*bursa.Wallet, *bursa.Wallet, error) {
	w1, _ := loadWallet("mywallet1")
	w2, _ := loadWallet("mywallet2")
	return w1, w2, nil
}

func loadWallet(walletName string) (*bursa.Wallet, error) {
	walletPath := GetConfig().GetWalletPath(walletName)
	wf, err := os.ReadFile(walletPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read wallet file %s", err.Error())
	}
	newWallet := bursa.Wallet{}
	json.Unmarshal(wf, &newWallet)
	return &newWallet, nil
}

func getUtxosByAddress(addr string) ([]UTxO.UTxO, error) {
	bfc, err := getBlockfrostContext()
	if err != nil {
		return nil, err
	}
	serAddr, err := serAddress.DecodeAddress(addr)
	if err != nil {
		return nil, err
	}
	utxos := bfc.Utxos(serAddr)
	if len(utxos) == 0 {
		return nil, fmt.Errorf("no utxos found for %s", addr)
	}

	return utxos, nil
}

func getBlockfrostContext() (*BlockFrostChainContext.BlockFrostChainContext, error) {
	cfg := GetConfig()
	var ret BlockFrostChainContext.BlockFrostChainContext
	switch cfg.Network {
	case "preprod":
		ret = BlockFrostChainContext.NewBlockfrostChainContext(
			constants.BLOCKFROST_BASE_URL_PREPROD,
			int(constants.PREPROD),
			cfg.BlockfrostApiKey,
		)
	default:
		return nil, fmt.Errorf("unsupported network: %s", cfg.Network)
	}
	return &ret, nil
}

func handleRequestTxIds(
	ctx txsubmission.CallbackContext,
	blocking bool,
	ack uint16,
	req uint16,
) ([]txsubmission.TxIdAndSize, error) {
	if ntnSentTx {
		// Terrible syncronization hack for shutdown
		close(ntnDoneChan)
		time.Sleep(5 * time.Second)
		return nil, nil
	}
	ret := []txsubmission.TxIdAndSize{
		{
			TxId: txsubmission.TxId{
				EraId: uint16(ntnTxType),
				TxId:  ntnTxHash,
			},
			Size: uint32(len(ntnTxBytes)),
		},
	}
	return ret, nil
}

func handleRequestTxs(
	ctx txsubmission.CallbackContext,
	txIds []txsubmission.TxId,
) ([]txsubmission.TxBody, error) {
	ret := []txsubmission.TxBody{
		{
			EraId:  uint16(ntnTxType),
			TxBody: ntnTxBytes,
		},
	}
	ntnSentTx = true
	return ret, nil
}
