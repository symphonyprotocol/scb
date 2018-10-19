package main

// import "fmt"
// import . "./block"
import . "./cli"

func main(){
	// tx := block.NewCoinbaseTX("xxx", "i love music")
	// iscoinbase := tx.IsCoinbase()
	// fmt.Print(iscoinbase)
	// bc := block.LoadBlockchain()
	// block.CreateBlockchain("trumpAddress")
	// block.GetBalance("trumpAddress")
	// bc.FindUTXO("trumpAddress")
	cli := CLI{}
	cli.Run()
	
}