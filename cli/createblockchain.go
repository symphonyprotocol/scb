package cli

import (
	"fmt"
	"github.com/symphonyprotocol/scb/block"
)

func (cli *CLI) CreateBlockchain(address, wif string) {
	flag := make(chan struct{})

	 block.CreateBlockchain(address, wif, func(bc *block.Blockchain){
		db := bc.GetDB()
		db.Close()
		block.ChangeBalance(address, block.Subsidy)
		flag <- struct{}{}
	})
	<-flag
	fmt.Println("Done!")
}