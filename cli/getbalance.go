package cli

import (
	"fmt"
	"log"
	block "github.com/symphonyprotocol/scb/block"
	"github.com/symphonyprotocol/sutil/elliptic"
)

func (cli *CLI) getBalance(address string) {
	keyHashed, valid := elliptic.LoadAddress(address)

	if !valid {
		log.Panic("ERROR: Address is not valid")
	}

	bc := block.LoadBlockchain()
	utxoset := block.UTXOSet{
		Blockchain : bc,
	}
	db := bc.GetDB()
	defer db.Close()


	balance := 0
	utxos := utxoset.FindUTXO(keyHashed)
	for _, out := range utxos{
		balance += out.Value
	}
	// UTXOs := bc.FindUTXO(keyHashed)
	// for _, out := range UTXOs {
	// 	balance += out.Value
	// }
	fmt.Printf("Balance of '%s': %d\n", address, balance)
}