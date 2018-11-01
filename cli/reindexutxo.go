package cli

import "fmt"
import "github.com/symphonyprotocol/scb/block"

func (cli *CLI) reindexUTXO() {
	bc := block.LoadBlockchain()
	utxoset := block.UTXOSet{
		Blockchain: bc,
	}
	utxoset.Reindex()
	count := utxoset.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}
