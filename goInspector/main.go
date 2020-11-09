package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		fmt.Println("Connect failed: ", err)
		return
	}

	flag := true
	var option int
	for flag {
		fmt.Println("Please choose your option:")
		fmt.Println("0: Check the balance of a certain account.\t1: Check transactions in a certain block.")
		fmt.Println("2: Check the tansaction with a certain hash.\t3: Start a transaction.")
		fmt.Println("4: Exit.")
		_, err = fmt.Scanln(&option)
		if err != nil {
			fmt.Println("Invalid input")
			continue
		}

		switch option {
		case 0:
			PrintBalance(client)
		case 1:
			PrintTransactionsInBlock(client)
		case 2:
			PrintTransactionByHash(client)
		case 3:
			SendTransaction(client)
		case 4:
			flag = false
		default:
			fmt.Println("Invalid input")
		}

		fmt.Println("")
	}
}
