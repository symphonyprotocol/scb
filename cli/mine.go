package cli
import "github.com/symphonyprotocol/scb/block"

func(cli *CLI) Mine(wif string){
	block.Mine(wif, func(transactions [] *block.Transaction){
		
	})

	// block.ChangeBalance(address, block.Subsidy)
	// for _, trans := range transactions{
	// 	block.ChangeBalance(trans.From, 0 - trans.Amount)
	// 	block.ChangeBalance(trans.To, trans.Amount)
	// }
}