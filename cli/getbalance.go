package cli

import (
	"github.com/symphonyprotocol/scb/block"
	"fmt"
)

func (cli *CLI) GetBalance(address string) (int64, int64){
	balance, gas := block.GetBalance(address)
	fmt.Printf("balancce is :%v\n", balance)
	fmt.Printf("gas is :%v\n", gas)
	return balance, gas
}