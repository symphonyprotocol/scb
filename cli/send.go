package cli

import (
	"github.com/symphonyprotocol/scb/block"
	"fmt"
)

func (cli *CLI) Send(from, to, wif string, amount int64){
	fmt.Printf("from :%s , to :%s, amount:%v\n", from , to, amount)
	block.SendTo(from, to, amount, wif)
}