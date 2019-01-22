package cli
import "github.com/symphonyprotocol/scb/block"
import "fmt"

func(cli *CLI) Mine(wif string){
	sign := make(chan struct{})
	block.Mine(wif, func (transactions [] *block.Transaction) {
		sign <- struct{}{}
		fmt.Print("done~")
	})

	<- sign
}