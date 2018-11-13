package cli

import (
	"fmt"
	"strconv"
	"github.com/symphonyprotocol/scb/block"
)

func (cli *CLI) printChain() {
	bc := block.LoadBlockchain()
	db := bc.GetDB()
	defer db.Close()

	bci := bc.Iterator()

	for {
		b := bci.Next()

		fmt.Printf("Prev. hash: %x\n", b.Header.PrevBlockHash)
		fmt.Printf("Hash: %x\n", b.Header.Hash)
		pow := block.NewProofOfWork(b)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(b.Header.PrevBlockHash) == 0 {
			break
		}
	}
}
