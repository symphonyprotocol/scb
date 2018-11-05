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
	utxoset := block.UTXOSet{
		Blockchain: bc,
	}

	db := bc.GetDB()
	defer db.Close()

	tx := block.NewUTXOTransaction(from, to, amount, &utxoset, prikey)
	fmt.Print(tx)
	//save to bolt db
	bc.SaveUTXOTransaction(tx)
	// cbtx := block.NewCoinbaseTX(from, "")
	// txs := []* block.Transaction{cbtx, tx}

	// newblock := bc.MineBlock(txs)
	// utxoset.Update(newblock)
	fmt.Println("Success!")
}