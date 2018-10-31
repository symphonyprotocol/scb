package cli

import (
	"fmt"
	"strconv"
	block "github.com/symphonyprotocol/scb/block"
)

func (cli *CLI) printChain() {
	bc := block.LoadBlockchain()
	db := bc.GetDB()
	defer db.Close()

	bci := bc.Iterator()

	for {
		b := bci.Next()

		fmt.Printf("Prev. hash: %x\n", b.PrevBlockHash)
		fmt.Printf("Hash: %x\n", b.Hash)
		pow := block.NewProofOfWork(b)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(b.PrevBlockHash) == 0 {
			break
		}
	}
}
