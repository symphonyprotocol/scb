package cli
import "github.com/symphonyprotocol/scb/block"

func(cli *CLI) Mine(address string){
	block.Mine(address)
	block.ChangeBalance(address, block.Subsidy)
}