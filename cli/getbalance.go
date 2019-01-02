package cli

import (
	"github.com/symphonyprotocol/scb/block"
	"fmt"
)

func (cli *CLI) GetBalance(address string) (int64){
	balance := block.GetBalance(address)
	fmt.Printf("balancce is :%v\n", balance)
	return balance
}