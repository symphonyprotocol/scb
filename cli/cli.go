package cli

import (
	"flag"
// 	"fmt"
	"log"
	"os"
// 	"strconv"
// 	block "github.com/symphonyprotocol/scb/block"
)

type CLI struct{}


func (cli *CLI) Run() {
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	// createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	createBlockchainWif := createBlockchainCmd.String("wif", "", "your wif private key")

	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int64("amount", 0, "Amount to send")
	wif := sendCmd.String("wif", "", "your wif private key")
	coinbase := sendCmd.String("coinbase", "", "if transaction is coinbase, empty will be false other will be true")

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")

	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	mineCmd := flag.NewFlagSet("mine", flag.ExitOnError)
	mineAddress := mineCmd.String("wif", "", "miner wif")


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
	case "mine":
		err := mineCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	}
	if createBlockchainCmd.Parsed() {
		// if *createBlockchainAddress == "" {
		// 	createBlockchainCmd.Usage()
		// 	os.Exit(1)
		// }
		cli.CreateBlockchain(*createBlockchainWif)
	}
	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}
		var flag bool
		if *coinbase == ""{
			flag = false
		}else{
			flag = true
		}
		cli.Send(*sendFrom, *sendTo,*wif, *sendAmount, flag)
	}
	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.GetBalance(*getBalanceAddress)
	}
	if printChainCmd.Parsed() {
		cli.PrintChain()
	}
	if mineCmd.Parsed(){
		if *mineAddress == ""{
			mineCmd.Usage()
			os.Exit(1)
		}
		cli.Mine(*mineAddress)
	}
}