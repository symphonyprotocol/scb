package cli

import (
	"fmt"
	"github.com/symphonyprotocol/scb/block"
)

func (cli *CLI) CreateBlockchain(wif string) {
	flag := make(chan struct{})

	 block.CreateBlockchain(wif, func(bc *block.Blockchain){
		fmt.Println("create block chain done")
		flag <- struct{}{}
	})
	<-flag
	fmt.Println("Done!")
}