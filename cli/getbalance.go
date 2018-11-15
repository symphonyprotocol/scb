package cli

import (
	"github.com/symphonyprotocol/scb/block"
)

func (cli *CLI) GetBalance(address string) int64{
	balance := block.GetBalance(address)
	return balance
}