package cli

import (
	"github.com/symphonyprotocol/scb/block"
	"fmt"
)

func (cli *CLI) Send(from, to, wif string, amount int64, coinbase bool){
	fmt.Printf("from :%s , to :%s, amount:%v, coinbase : %v", from , to, amount, coinbase)
	block.SendTo(from, to, amount, wif, coinbase)
}