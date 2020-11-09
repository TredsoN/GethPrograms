package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func PrintBalance(client *ethclient.Client) {
	var accountstr string

	fmt.Println("Please input your account address:")
	fmt.Scanln(&accountstr)

	account := common.HexToAddress(accountstr)

	balance, err := client.BalanceAt(context.Background(), account, nil)
	if err != nil {
		fmt.Println("Get balance failed: ", err)
		return
	}

	fmt.Println("Balance: ", balance)
}

func PrintTransactionsInBlock(client *ethclient.Client) {
	var blockNum int64
	fmt.Println("Please input the block id:")
	_, err := fmt.Scanln(&blockNum)
	if err != nil {
		fmt.Println("Invalid input.")
		return
	}

	blockNumber := big.NewInt(blockNum)
	block, err := client.BlockByNumber(context.Background(), blockNumber)
	if err != nil {
		fmt.Println("Get transaction failed: ", err)
		return
	}

	if len(block.Transactions()) == 0 {
		fmt.Println("No transactions!")
		return
	}

	for idx, tx := range block.Transactions() {
		fmt.Printf("Transaction %v:\n{\n", idx)
		fmt.Printf("  Hash: %s,\n", tx.Hash().Hex())
		fmt.Printf("  Value: %s,\n", tx.Value().String())
		fmt.Printf("  Gas: %v,\n", tx.Gas())
		fmt.Printf("  GasPrice: %v,\n", tx.GasPrice().Uint64())
		fmt.Printf("  Nonce: %v,\n", tx.Nonce())
		fmt.Printf("  To: %s,\n", tx.To().Hex())

		chainID, err := client.NetworkID(context.Background())
		if err != nil {
			fmt.Println("Get transaction failed: ", err)
			return
		}

		if msg, err := tx.AsMessage(types.NewEIP155Signer(chainID)); err == nil {
			fmt.Printf("  From: %s,\n", msg.From().Hex())
		}

		receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
		if err != nil {
			fmt.Println("Get transaction failed: ", err)
			return
		}

		fmt.Printf("  Status: %v,\n}\n", receipt.Status)
	}
}

func PrintTransactionByHash(client *ethclient.Client) {
	var hashStr string
	fmt.Println("Please input the hash of transation:")
	fmt.Scanln(&hashStr)

	txHash := common.HexToHash(hashStr)
	tx, isPending, err := client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		fmt.Println("Get transaction failed: ", err)
		return
	}

	fmt.Printf("{\n  Hash: %s,\n", tx.Hash().Hex())
	fmt.Printf("  Value: %s,\n", tx.Value().String())
	fmt.Printf("  Gas: %v,\n", tx.Gas())
	fmt.Printf("  GasPrice: %v,\n", tx.GasPrice().Uint64())
	fmt.Printf("  Nonce: %v,\n", tx.Nonce())
	fmt.Printf("  To: %s,\n", tx.To().Hex())

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		fmt.Println("Get transaction failed: ", err)
		return
	}

	if msg, err := tx.AsMessage(types.NewEIP155Signer(chainID)); err == nil {
		fmt.Printf("  From: %s,\n", msg.From().Hex())
	}

	fmt.Printf("  IsPending: %v,\n}\n", isPending)
}

func SendTransaction(client *ethclient.Client) {
	var privateKeyFile, password string
	fmt.Println("Please input your key file path:")
	fmt.Scanln(&privateKeyFile)
	fmt.Println("Please input your password:")
	fmt.Scanln(&password)

	var keyValue = GetPrivateKey(&privateKeyFile, &password)

	privateKey, err := crypto.HexToECDSA(keyValue)
	if err != nil {
		fmt.Println("Get privateKey failed: ", err)
		return
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		fmt.Println("Error casting public key to ECDSA.")
		return
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	var nonceStr string
	var nonce uint64
	fmt.Println("(Transaction Config) Please input nonce (if skipped, nonce will be set as the default):")
	_, err = fmt.Scanln(&nonceStr)
	if err != nil {
		nonceStr = ""
	}

	if nonceStr == "" {
		nonce, err = client.PendingNonceAt(context.Background(), fromAddress)
		if err != nil {
			fmt.Println("Send transaction failed: ", err)
			return
		}
	} else {
		intNum, err := strconv.Atoi(nonceStr)
		if err != nil {
			fmt.Println("Invalid input.")
			return
		}
		nonce = uint64(intNum)
	}

	var valueStr string
	fmt.Println("(Transaction Config) Please input value:")
	fmt.Scanln(&valueStr)
	value := new(big.Int)
	value, ok = value.SetString(valueStr, 10)
	if !ok {
		fmt.Println("Invalid input.")
		return
	}

	var gasLimitInt int
	fmt.Println("(Transaction Config) Please input gas limit:")
	_, err = fmt.Scanln(&gasLimitInt)
	if err != nil {
		fmt.Println("Invalid input.")
		return
	}
	gasLimit := uint64(gasLimitInt)

	var gasPriceStr string
	fmt.Println("(Transaction Config) Please input gas price:")
	fmt.Scanln(&gasPriceStr)
	gasPrice := new(big.Int)
	gasPrice, ok = gasPrice.SetString(gasPriceStr, 10)

	if !ok {
		fmt.Println("Invalid input.")
		return
	}

	var accountstr string
	fmt.Println("(Transaction Config) Please input recipient account address:")
	fmt.Scanln(&accountstr)
	toAddress := common.HexToAddress(accountstr)

	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		fmt.Println("Send transaction failed: ", err)
		return
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		fmt.Println("Send transaction failed: ", err)
		return
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		fmt.Println("Send transaction failed: ", err)
		return
	}

	fmt.Println("Transaction has been sent, hash: ", signedTx.Hash().Hex())
}

func GetPrivateKey(privateKeyFile, password *string) string {
	keyJSON, err := ioutil.ReadFile(*privateKeyFile)
	if err != nil {
		fmt.Println("Get privateKey failed: ", err)
		return ""
	}

	unlockedKey, err := keystore.DecryptKey(keyJSON, *password)
	if err != nil {
		fmt.Println("Get privateKey failed: ", err)
		return ""
	}

	privKey := hex.EncodeToString(unlockedKey.PrivateKey.D.Bytes())

	return privKey
}
