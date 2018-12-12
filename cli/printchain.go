package cli

import (
	"github.com/symphonyprotocol/scb/block"
)

func (cli *CLI) PrintChain() {
	block.PrintChain()
}
