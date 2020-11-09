package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
	"github.com/tyler-smith/go-bip39"
)

// GenerateAccount : returns the mnemontic of the user
func GenerateAccount() (*string, error) {
	// generate a new mnemonic
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return nil, errors.New("fail to generate mnemonic:" + err.Error())
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return nil, errors.New("fail to generate mnemonic:" + err.Error())
	}

	// generate a wallet
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, errors.New("fail to generate account:" + err.Error())
	}

	path := hdwallet.MustParseDerivationPath("m/44'/60'/0'/0/0")
	account, err := wallet.Derive(path, false)
	if err != nil {
		return nil, errors.New("fail to generate account:" + err.Error())
	}

	// store key
	ks := keystore.NewKeyStore("../data/keystore", keystore.StandardScryptN, keystore.StandardScryptP)

	privateKey, err := wallet.PrivateKey(account)
	if err != nil {
		return nil, errors.New("fail to generate account:" + err.Error())
	}

	_, err = ks.ImportECDSA(privateKey, "admin")
	if err != nil {
		return nil, errors.New("fail to generate account:" + err.Error())
	}

	return &mnemonic, nil
}

// RegisterAnAccount : returns the address and mnemonic of an available account
func RegisterAnAccount() (*string, *string, error) {
	// open pool file
	accountdataptr, err := ReadFileContent(AcountPoolPath)
	if err != nil {
		return nil, nil, errors.New("fail to open account pool file: " + err.Error())
	}

	accountdata := *accountdataptr

	accounts, ok := accountdata["accounts"].([]interface{})
	if !ok {
		return nil, nil, errors.New("fail to open account pool file")
	}

	// get an available address
	var tgaddress, tgmnemonic string
	newaccounts := []interface{}{}
	found := false
	for _, v := range accounts {
		account, ok := v.(map[string]interface{})
		if !ok {
			return nil, nil, errors.New("fail to open account pool file")
		}

		if account["status"] == "0" && !found {
			found = true
			account["status"] = "1"
			tgaddress, _ = account["address"].(string)
			tgmnemonic, _ = account["mnemonic"].(string)
		}

		newaccounts = append(newaccounts, account)
	}

	if !found {
		return nil, nil, errors.New("no available account")
	}

	accountdata["accounts"] = newaccounts

	newaccountdata, err := json.MarshalIndent(accountdata, "", "")
	if err != nil {
		return nil, nil, errors.New("fail to write account pool file")
	}

	// rewrite account info
	file, err := os.Create(AcountPoolPath)
	if err != nil {
		return nil, nil, errors.New("fail to write account pool file: " + err.Error())
	}

	defer file.Close()

	_, err = file.WriteString(string(newaccountdata))
	if err != nil {
		return nil, nil, errors.New("fail to write account pool file: " + err.Error())
	}

	return &tgaddress, &tgmnemonic, nil
}

// LoginAnAccount : log in the system
func LoginAnAccount(mnemonic string) bool {
	// get address
	address, err := MnemonicToAccount(mnemonic)
	if err != nil {
		return false
	}

	curuser = strings.ToLower(*address)
	return true
}

// RefreshAccountInfo : monitor transactions by hash and update current user's information file
// returns pending balance
func RefreshAccountInfo(client *ethclient.Client, address string) (*string, *string, error) {
	// read info file content
	dataptr, err := ReadFileContent(AccountInfoPath + address + ".txt")
	if err != nil {
		return nil, nil, errors.New("fail to open account info file: " + err.Error())
	}

	data := *dataptr

	// get balance, addrbalance and transactions
	var balance, addrbalance, pendingbalance *big.Int
	if value, ok := data["balance"].(string); ok {
		balance = new(big.Int)
		balance, _ = balance.SetString(value, 10)
	}
	if value, ok := data["addrbalance"].(string); ok {
		addrbalance = new(big.Int)
		addrbalance, _ = addrbalance.SetString(value, 10)
	}
	if value, ok := data["pendingbalance"].(string); ok {
		pendingbalance = new(big.Int)
		pendingbalance, _ = pendingbalance.SetString(value, 10)
	}
	var transactions []interface{}
	if value, ok := data["transactions"].([]interface{}); ok {
		transactions = value
	}

	// check all transaction by hash and refresh account info
	var newtransactions []interface{}
	haspending := false
	for _, t := range transactions {
		txmap, _ := t.(map[string]interface{})
		tx, err := MapToTransaction(txmap)
		if err != nil {
			return nil, nil, errors.New("fail to get transaction: " + err.Error())
		}

		if tx.Status == "0" {
			// get transaction by hash
			txHash := common.HexToHash(tx.Hash)
			_, ispending, err := client.TransactionByHash(context.Background(), txHash)
			if err != nil {
				return nil, nil, errors.New("fail to get transaction: " + err.Error())
			}

			// refresh transaction status
			if !ispending {
				tx.Status = "1"
				switch tx.Type {
				case "0":
					amount := new(big.Int)
					amount, ok := amount.SetString(tx.Amount, 10)
					if !ok {
						return nil, nil, errors.New("fail to get amount")
					}
					balance.Add(balance, amount)
					addrbalance.Add(addrbalance, amount)
				case "1":
					amount := new(big.Int)
					amount, ok := amount.SetString(tx.Amount, 10)
					if !ok {
						return nil, nil, errors.New("fail to get amount")
					}
					balance.Sub(balance, amount)
					addrbalance.Sub(addrbalance, amount)
				case "2":
					amount := new(big.Int)
					amount, ok := amount.SetString(tx.Amount, 10)
					if !ok {
						return nil, nil, errors.New("fail to get amount")
					}
					addrbalance.Sub(addrbalance, amount)
				}
			} else {
				haspending = true
				if tx.Type == "1" {
					amount := new(big.Int)
					amount, ok := amount.SetString(tx.Amount, 10)
					if !ok {
						return nil, nil, errors.New("fail to get amount")
					}
					pendingbalance = pendingbalance.Sub(balance, amount)
				}
			}
			newtransactions = append(newtransactions, TransactionToMap(tx))
		} else {
			newtransactions = append(newtransactions, txmap)
		}
	}

	// sychronize balance and pending balance
	if !haspending {
		pendingbalance = balance
	}

	// generate new info
	newinfo := make(map[string]interface{})
	newinfo["balance"] = balance.String()
	newinfo["addrbalance"] = addrbalance.String()
	newinfo["pendingbalance"] = pendingbalance.String()
	newinfo["transactions"] = newtransactions

	newinfodata, err := json.MarshalIndent(newinfo, "", "")
	if err != nil {
		return nil, nil, errors.New("fail to write user file")
	}

	// save new data
	file, err := os.Create(AccountInfoPath + address + ".txt")
	if err != nil {
		return nil, nil, errors.New("fail to write user file: " + err.Error())
	}

	defer file.Close()

	_, err = file.WriteString(string(newinfodata))
	if err != nil {
		return nil, nil, errors.New("fail to write userfile: " + err.Error())
	}

	adbl := addrbalance.String()
	pbl := pendingbalance.String()
	return &adbl, &pbl, nil
}

// RefreshAllAccount : monitor transactions by hash and update all users' information file
func RefreshAllAccount(client *ethclient.Client) (*[]map[string]string, bool) {
	// open pool file
	accountdataptr, err := ReadFileContent(AcountPoolPath)
	if err != nil {
		return nil, false
	}

	accountdata := *accountdataptr

	accounts, ok := accountdata["accounts"].([]interface{})
	if !ok {
		return nil, false
	}

	// calculate all the balance
	info := []map[string]string{}
	for _, v := range accounts {
		userinfo := make(map[string]string)
		account, _ := v.(map[string]interface{})

		if acctype, _ := account["status"].(string); acctype != "1" {
			continue
		}

		address, _ := account["address"].(string)

		addressbalance, _, err := RefreshAccountInfo(client, address)
		if err != nil {
			fmt.Println(err)
			return nil, false
		}
		userinfo["address"] = address
		userinfo["addressbalance"] = *addressbalance

		info = append(info, userinfo)
	}

	return &info, true
}

// StartATransaction : start a transaction, returns the hash
func StartATransaction(client *ethclient.Client, value *big.Int, from, to string, privateKey *ecdsa.PrivateKey) (*string, error) {
	// generate transaction
	gasLimit := uint64(80000)

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, err
	}

	delta := new(big.Int)
	delta, _ = delta.SetString("5000000000", 10)
	gasPrice = gasPrice.Add(gasPrice, delta)

	fromAddress := common.HexToAddress(from)
	toAddress := common.HexToAddress(to)

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, err
	}

	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	// get chain id
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, err
	}

	// sign the transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return nil, err
	}

	// send transactions
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, err
	}

	hash := signedTx.Hash().Hex()

	return &hash, nil
}

// SaveATransaction : save a transaction to account info file
func SaveATransaction(address, hash, tp, amount string) error {
	// read account info
	dataptr, err := ReadFileContent(AccountInfoPath + address + ".txt")
	if err != nil {
		return errors.New(("fail to open account info file: " + err.Error()))
	}

	data := *dataptr

	// get balance, addrbalance and transactions
	var balance, addrbalance, pendingbalance *big.Int
	if value, ok := data["balance"].(string); ok {
		balance = new(big.Int)
		balance, _ = balance.SetString(value, 10)
	}
	if value, ok := data["addrbalance"].(string); ok {
		addrbalance = new(big.Int)
		addrbalance, _ = addrbalance.SetString(value, 10)
	}
	if value, ok := data["pendingbalance"].(string); ok {
		pendingbalance = new(big.Int)
		pendingbalance, _ = pendingbalance.SetString(value, 10)
	}
	var transactions []interface{}
	if value, ok := data["transactions"].([]interface{}); ok {
		transactions = value
	}

	transaction := MyTransaction{Hash: hash, Type: tp, Status: "0", Amount: amount}
	transactions = append(transactions, TransactionToMap(&transaction))

	// generate new info
	newinfo := make(map[string]interface{})
	newinfo["balance"] = balance.String()
	newinfo["addrbalance"] = addrbalance.String()
	newinfo["pendingbalance"] = pendingbalance.String()
	newinfo["transactions"] = transactions

	newinfodata, err := json.MarshalIndent(newinfo, "", "")
	if err != nil {
		return errors.New(("fail to write user file: " + err.Error()))
	}

	// save new data
	file, err := os.Create(AccountInfoPath + address + ".txt")
	if err != nil {
		return errors.New(("fail to write user file: " + err.Error()))
	}

	defer file.Close()

	_, err = file.WriteString(string(newinfodata))
	if err != nil {
		return errors.New(("fail to write user file: " + err.Error()))
	}

	return nil
}
