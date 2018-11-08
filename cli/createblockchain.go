package cli

import (
	"fmt"
	// "log"
	// "github.com/symphonyprotocol/sutil/elliptic"
	block "github.com/symphonyprotocol/scb/block"
)

func (cli *CLI) CreateBlockchain(address string) {
	// _, valid := elliptic.LoadAddress(address)

	// if !valid {
	// 	log.Panic("ERROR: Address is not valid")
	// }
	bc := block.CreateBlockchain(address)
	db := bc.GetDB()
	defer db.Close()

	// 重建utxo集
	utxoset := block.UTXOSet{
		Blockchain: bc,
	}
	utxoset.Reindex()

	fmt.Println("Done!")
}