package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/ethclient"
)

var curuser string
var keystoremap map[string]string

func main() {
	keystoremap = make(map[string]string)
	keystoremap["0xc0093215bec3cbb9522352dcb4e3fa8fd5b665d1"] = `D:\Projects\GethProgram\chain_10\keystore\UTC--2020-07-30T12-08-00.423453600Z--c0093215bec3cbb9522352dcb4e3fa8fd5b665d1`
	keystoremap["0x8bf8fb9cfd048de2645539f5007d12ae6465623b"] = `D:\Projects\GethProgram\chain_10\keystore\UTC--2020-07-31T01-19-22.030531400Z--8bf8fb9cfd048de2645539f5007d12ae6465623b`
	keystoremap["0x297fca8450553a3c9103dd410b59e79cdf6f4a0a"] = `D:\Projects\GethProgram\chain_10\keystore\UTC--2020-07-31T01-19-25.477316000Z--297fca8450553a3c9103dd410b59e79cdf6f4a0a`
	keystoremap["0x85bdd4c54fc8689e2973ab8ded1cf46c55243e05"] = `D:\Projects\GethProgram\chain_10\keystore\UTC--2020-07-31T01-19-28.576035000Z--85bdd4c54fc8689e2973ab8ded1cf46c55243e05`
	keystoremap["0x7e5befa8b17710f881f9011425a93e47dfa8590e"] = `D:\Projects\GethProgram\chain_10\keystore\UTC--2020-07-31T01-19-32.013843100Z--7e5befa8b17710f881f9011425a93e47dfa8590e`
	keystoremap["0xf81233b607aa4178ed9e9e1e1a876b683619ef47"] = `D:\Projects\GethProgram\chain_10\keystore\UTC--2020-07-31T01-19-35.118545500Z--f81233b607aa4178ed9e9e1e1a876b683619ef47`

	var option int
	islogin := false
	isadmin := false

	for !islogin {
		fmt.Println("please choose your option:")
		fmt.Println("0: log in\t1: register\t2: log in as admin\t3: exit")

		_, err := fmt.Scanln(&option)
		if err != nil {
			fmt.Println("invalid input")
			continue
		}

		switch option {
		case 0:
			islogin = Login()
			fmt.Println("")
		case 1:
			Register()
			fmt.Println("")
		case 2:
			if LoginAsAdmin() {
				islogin = true
				isadmin = true
			}
			fmt.Println("")
		case 3:
			fmt.Println("")
			return
		default:
			fmt.Println("invalid input")
			fmt.Println("")
		}
	}

	client, err := ethclient.Dial(RPCAddress)
	if err != nil {
		fmt.Println("connect failededed: ", err)
		return
	}

	for !isadmin {
		fmt.Println("please choose your option:")
		fmt.Println("0: check balance\t1: recharge\t2: withdraw\t3: exit")

		_, err := fmt.Scanln(&option)
		if err != nil {
			fmt.Println("invalid input")
			continue
		}

		switch option {
		case 0:
			CheckBalance(client)
			fmt.Println("")
		case 1:
			Recharge(client)
			fmt.Println("")
		case 2:
			Withdraw(client)
			fmt.Println("")
		case 3:
			fmt.Println("")
			return
		default:
			fmt.Println("invalid input")
			fmt.Println("")
		}
	}

	for isadmin {
		fmt.Println("please choose your option:")
		fmt.Println("0: centralize\t1: exit")

		_, err := fmt.Scanln(&option)
		if err != nil {
			fmt.Println("invalid input")
			continue
		}

		switch option {
		case 0:
			Centralize(client)
			fmt.Println("")
		case 1:
			fmt.Println("")
			return
		default:
			fmt.Println("invalid input")
			fmt.Println("")
		}
	}
}
