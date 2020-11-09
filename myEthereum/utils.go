package main

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

// MyTransaction : a transaction on the chain
// transaction type
//  0: recharge
//  1: withdraw
//  2: centralize
// transaction statuse
//  0: pending
//  1: done
type MyTransaction struct {
	Hash   string `json:"hash"`
	Type   string `json:"type"`
	Status string `json:"status"`
	Amount string `json:"amount"`
}

const (
	// AcountPoolPath : the file storing account pool information
	AcountPoolPath = `.\SystemData\addresses.txt`
	// AccountInfoPath : the file storing account information
	AccountInfoPath = `.\SystemData\AccountInfo\`
	// RPCAddress : the rpc address of the chain
	RPCAddress = "http://localhost:8545"
	// MainAddress : main address of the system
	MainAddress = "0xc0093215bec3cbb9522352dcb4e3fa8fd5b665d1"
)

// ReadFileContent : returns the file content as json
func ReadFileContent(path string) (*map[string]interface{}, error) {
	// open file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	// read file content
	buf := make([]byte, 1024)
	data := make([]byte, 0)

	for {
		n, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}

		data = append(data, buf[:n]...)
	}

	var datajson map[string]interface{}
	err = json.Unmarshal(data, &datajson)
	if err != nil {
		return nil, err
	}
	return &datajson, nil
}

// DecomposeTransaction : returns a MyTransaction
func DecomposeTransaction(client *ethclient.Client, tx *types.Transaction, tp string) (*MyTransaction, error) {
	hash := tx.Hash().Hex()
	amount := tx.Value().String()

	return &MyTransaction{Hash: strings.ToLower(hash), Amount: amount, Type: tp, Status: "0"}, nil
}

// MapToTransaction : transfer a map to a MyTransactions
func MapToTransaction(txmap map[string]interface{}) (*MyTransaction, error) {
	var tp, status, hash, amount string
	if value, ok := txmap["type"].(string); ok {
		tp = value
	}
	if value, ok := txmap["status"].(string); ok {
		status = value
	}
	if value, ok := txmap["hash"].(string); ok {
		hash = value
	}
	if value, ok := txmap["amount"].(string); ok {
		amount = value
	}
	if tp != "" && status != "" && hash != "" && amount != "" {
		return &MyTransaction{
			Hash:   hash,
			Type:   tp,
			Status: status,
			Amount: amount,
		}, nil
	}
	return nil, errors.New("wrong map format")
}

// TransactionToMap : transfer a MyTransactions to a map
func TransactionToMap(tx *MyTransaction) map[string]interface{} {
	txmap := make(map[string]interface{})
	txmap["hash"] = tx.Hash
	txmap["type"] = tx.Type
	txmap["status"] = tx.Status
	txmap["amount"] = tx.Amount
	return txmap
}

// MnemonicToAccount : returns the account address from a mnemonic.
func MnemonicToAccount(mnemonic string) (*string, error) {
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, errors.New("failed to get account:" + err.Error())
	}

	path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := wallet.Derive(path, false)
	if err != nil {
		fmt.Println("failed to generate wallet:", err)
		return nil, errors.New("failed to get account:" + err.Error())
	}

	address := account.Address.Hex()
	return &address, nil
}

// GetPrivateKey : get privatekey from keystore
func GetPrivateKey(privateKeyFile, password *string) (*string, error) {
	keyJSON, err := ioutil.ReadFile(*privateKeyFile)
	if err != nil {
		return nil, err
	}

	unlockedKey, err := keystore.DecryptKey(keyJSON, *password)
	if err != nil {
		return nil, err
	}

	privKey := hex.EncodeToString(unlockedKey.PrivateKey.D.Bytes())

	return &privKey, nil
}
