package cli
import "github.com/symphonyprotocol/scb/block"

func(cli *CLI) Mine(address string){
	transactions := block.Mine(address)
	block.ChangeBalance(address, block.Subsidy)
	for _, trans := range transactions{
		block.ChangeBalance(trans.From, 0 - trans.Amount)
		block.ChangeBalance(trans.To, trans.Amount)
	}
}