package cli

import (
	"flag"
	"fmt"
	"log"
	"os"
	// "strconv"
	block "../block"
)

type CLI struct{}

func (cli *CLI) createBlockchain(address string) {
	bc := block.CreateBlockchain(address)
	db := bc.GetDB()
	db.Close()
	fmt.Println("Done!")
}

func (cli *CLI) Run() {
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	switch os.Args[1] {
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
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
}