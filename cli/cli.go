package cli

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	block "../block"
)

type CLI struct{}

func (cli *CLI) createBlockchain(address string) {
	bc := block.CreateBlockchain(address)
	db := bc.GetDB()
	defer db.Close()
	fmt.Println("Done!")
}
func (cli *CLI) send(from, to string, amount int) {
	bc := block.LoadBlockchain()
	db := bc.GetDB()
	defer db.Close()

	tx := block.NewUTXOTransaction(from, to, amount, bc)
	bc.MineBlock([]* block.Transaction{tx})
	fmt.Println("Success!")
}
func (cli *CLI) getBalance(address string) {
	bc := block.LoadBlockchain()
	db := bc.GetDB()
	defer db.Close()

	balance := 0
	UTXOs := bc.FindUTXO(address)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}
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


func (cli *CLI) Run() {
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")

	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")

	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	switch os.Args[1] {
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	}
	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress)
	}
	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}

		cli.send(*sendFrom, *sendTo, *sendAmount)
	}
	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}
	if printChainCmd.Parsed() {
		cli.printChain()
	}
}