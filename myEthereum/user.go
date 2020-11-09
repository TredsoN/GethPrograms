package main

import (
	"bufio"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Login : user log in
func Login() bool {
	fmt.Println("please input your mnemonic:")
	reader := bufio.NewReader(os.Stdin)
	mnemonic, _, _ := reader.ReadLine()

	if LoginAnAccount(string(mnemonic)) {
		fmt.Println("welcome, " + curuser)
		return true
	}

	fmt.Println("failed to log in: wrong mnemonic")
	return false
}

// LoginAsAdmin : log in as an admin
func LoginAsAdmin() bool {
	var password string
	fmt.Println("please input admin password:")
	fmt.Scanln(&password)

	if password != "admin" {
		fmt.Println("wrong password")
		return false
	}

	return true
}

// Register : user registers
func Register() {
	_, mnemonic, err := RegisterAnAccount()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("succeeded to register, your mnemonic is:\n%s\n", *mnemonic)
}

// CheckBalance : check the balance in current account
func CheckBalance(client *ethclient.Client) {
	_, balance, err := RefreshAccountInfo(client, curuser)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("your balance is:\n%s\n", *balance)
}

// Recharge : recharge
func Recharge(client *ethclient.Client) {
	// get privateKey
	var privateKeyFile, password string
	fmt.Println("please input your key file path:")
	fmt.Scanln(&privateKeyFile)
	fmt.Println("please input your password:")
	fmt.Scanln(&password)

	keyValue, err := GetPrivateKey(&privateKeyFile, &password)
	if err != nil {
		fmt.Println("failed to get privateKey: ", err)
		return
	}

	privateKey, err := crypto.HexToECDSA(*keyValue)
	if err != nil {
		fmt.Println("failed to get privateKey: ", err)
		return
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		fmt.Println("Error casting public key to ECDSA.")
		return
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	ethaddress := fromAddress.Hex()

	// get value
	var valuestr string
	fmt.Println("please input the value:")
	fmt.Scanln(&valuestr)

	value := new(big.Int)
	value, ok = value.SetString(valuestr, 10)
	if !ok {
		fmt.Println("invalid input")
		return
	}

	// send transaction
	hash, err := StartATransaction(client, value, ethaddress, curuser, privateKey)
	if err != nil {
		fmt.Println("fail to recharge: ", err)
		return
	}

	// save transaction
	err = SaveATransaction(curuser, *hash, "0", valuestr)
	if err != nil {
		fmt.Println("fail to save transaction: ", err)
		return
	}

	fmt.Printf("transaction submitted:%v\n", *hash)
}

// Withdraw : Withdraw
func Withdraw(client *ethclient.Client) {
	// get privateKey
	mk := keystoremap[strings.ToLower(MainAddress)]
	mp := "admin"
	keyValue, err := GetPrivateKey(&mk, &mp)
	if err != nil {
		fmt.Println("failed to get privateKey: ", err)
		return
	}

	privateKey, err := crypto.HexToECDSA(*keyValue)
	if err != nil {
		fmt.Println("failed to get privateKey: ", err)
		return
	}

	// get address
	var ethaddress string
	fmt.Println("please input your ethereum address:")
	fmt.Scanln(&ethaddress)

	// get value
	var valuestr string
	fmt.Println("please input the value:")
	fmt.Scanln(&valuestr)

	value := new(big.Int)
	value, ok := value.SetString(valuestr, 10)
	if !ok {
		fmt.Println("invalid input")
		return
	}

	// get balance
	_, balancestr, err := RefreshAccountInfo(client, curuser)
	if err != nil {
		fmt.Println(err)
		return
	}

	balance := new(big.Int)
	balance, ok = balance.SetString(*balancestr, 10)
	if !ok {
		fmt.Println("invalid input")
		return
	}

	// compare balance to value
	if balance.Cmp(value) == -1 {
		fmt.Println("balance is not enough")
		return
	}

	// send transaction
	hash, err := StartATransaction(client, value, MainAddress, ethaddress, privateKey)
	if err != nil {
		fmt.Println("fail to recharge: ", err)
		return
	}

	// save transaction
	err = SaveATransaction(curuser, *hash, "1", valuestr)
	if err != nil {
		fmt.Println("fail to save transaction: ", err)
		return
	}

	fmt.Printf("transaction submitted: %v\n", *hash)
}

// Centralize : centralize all the balance in user accounts
func Centralize(client *ethclient.Client) {
	// refresh all accounts
	info, ok := RefreshAllAccount(client)
	if !ok {
		fmt.Println("failed to refresh balance")
		return
	}

	fmt.Println("refresh accounts done")

	// send transactions
	for _, userinfo := range *info {
		// if address balance is 0, then continue
		if userinfo["addressbalance"] == "0" {
			continue
		}

		// get privateKey
		mk := keystoremap[strings.ToLower(userinfo["address"])]
		mp := "admin"

		keyValue, err := GetPrivateKey(&mk, &mp)
		if err != nil {
			fmt.Println("failed to get privateKey: ", err)
			return
		}

		privateKey, err := crypto.HexToECDSA(*keyValue)
		if err != nil {
			fmt.Println("failed to get privateKey: ", err)
			return
		}

		// get value
		value := new(big.Int)
		value, ok := value.SetString(userinfo["addressbalance"], 10)
		if !ok {
			fmt.Println("invalid input")
			return
		}

		// send transaction
		hash, err := StartATransaction(client, value, userinfo["address"], MainAddress, privateKey)
		if err != nil {
			fmt.Println("failed to centalize: ", err)
			return
		}

		// save transaction
		err = SaveATransaction(userinfo["address"], *hash, "2", userinfo["pendingbalance"])
		if err != nil {
			fmt.Println("fail to save transaction: ", err)
			return
		}

		fmt.Printf("[%s] transaction submitted: %v\n", userinfo["address"], *hash)
	}
}
