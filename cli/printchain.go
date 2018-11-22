package cli

import (
	"fmt"
	"strconv"
	"github.com/symphonyprotocol/scb/block"
)

func (cli *CLI) printChain() {
	bc := block.LoadBlockchain()
	bci := bc.Iterator()

	for {
		b := bci.Next()

		fmt.Printf("Previous hash: %x\n", b.Header.PrevBlockHash)
		fmt.Printf("Hash: %x\n", b.Header.Hash)
		fmt.Printf("CreateAt: %v\n", b.Header.Timestamp)
		fmt.Printf("Height:%d\n", b.Header.Height)
		fmt.Printf("Coinbase:%v\n", b.Header.Coinbase)
		fmt.Printf("merkle Root:%v\n", b.Header.MerkleRootHash)
		pow := block.NewProofOfWork(b)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Printf("Signature Verify:%v \n", b.VerifyCoinbase())
		fmt.Println()

		if len(b.Header.PrevBlockHash) == 0 {
			break
		}
	}
}
