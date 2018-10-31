package cli

import (
	"fmt"
	"log"
	"github.com/symphonyprotocol/sutil/elliptic"
	"github.com/symphonyprotocol/scb/block"
)

func (cli *CLI) send(from, to string, amount int, wif string) {
	_, validFrom := elliptic.LoadAddress(from)
	_, validTo := elliptic.LoadAddress(to)
	prikey, _ := elliptic.LoadWIF(wif)
	
	if !validFrom{
		log.Panic("ERROR: Sender address is not valid")
	}
	if !validTo{
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := block.LoadBlockchain()
	db := bc.GetDB()
	defer db.Close()

	tx := block.NewUTXOTransaction(from, to, amount, bc, prikey)
	bc.MineBlock([]* block.Transaction{tx})
	fmt.Println("Success!")
}