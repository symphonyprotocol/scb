package cli

import (
	"github.com/symphonyprotocol/scb/block"
)

func (cli *CLI) Send(from, to, wif string, amount int64, coinbase bool){
	block.SendTo(from, to, amount, wif, coinbase)
}