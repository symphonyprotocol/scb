package main

// import "fmt"
// import . "./block"


// import . "./cli"

import "github.com/symphonyprotocol/scb/cli"

// import "github.com/symphonyprotocol/sutil/elliptic"
// import "log"
// import "github.com/symphonyprotocol/scb/block"

func main(){
	// tx := block.NewCoinbaseTX("xxx", "i love music")
	// iscoinbase := tx.IsCoinbase()
	// fmt.Print(iscoinbase)
	// bc := block.LoadBlockchain()
	// block.CreateBlockchain("trumpAddress")
	// block.GetBalance("trumpAddress")
	// bc.FindUTXO("trumpAddress")



	cli := cli.CLI{}
	cli.Run()

	//1.create chain

	// bc := block.CreateBlockchain("1T3r9yFFM6St9wGSp7zMYP24G6pUYnL7y")
	// db := bc.GetDB()
	// defer db.Close()


	// 2. get balance.

	// keyHashed, valid := elliptic.LoadAddress("1T3r9yFFM6St9wGSp7zMYP24G6pUYnL7y")

	// if !valid {
	// 	log.Panic("ERROR: Address is not valid")
	// }

	// bc := block.LoadBlockchain()
	// db := bc.GetDB()
	// defer db.Close()

	// balance := 0
	// UTXOs := bc.FindUTXO(keyHashed)
	// for _, out := range UTXOs {
	// 	balance += out.Value
	// }


	// send coin
	
	// from := "1T3r9yFFM6St9wGSp7zMYP24G6pUYnL7y"
	// to := "189wh8VjXLmKSZhnP9DQwcVKfvNemQSmBp"
	// wif := "L5fR7FRHnZGL3DjsrhN8CvBYHpywL8LjxA2rjzbL7qvFqjgbNVQ5"
	// amount := 1

	// _, validFrom := elliptic.LoadAddress(from)
	// _, validTo := elliptic.LoadAddress(to)
	// prikey, _ := elliptic.LoadWIF(wif)

	// if !validFrom{
	// 	log.Panic("ERROR: Sender address is not valid")
	// }
	// if !validTo{
	// 	log.Panic("ERROR: Recipient address is not valid")
	// }

	// bc := block.LoadBlockchain()
	// db := bc.GetDB()
	// defer db.Close()

	// tx := block.NewUTXOTransaction(from, to, amount, bc, prikey)
	// bc.MineBlock([]* block.Transaction{tx})
	// fmt.Println("Success!")






	// fmt.Printf("Balance of '%s': %d\n", address, balance)

	




	// a := make(map[string][]int)

	// // a["1"]  = append(a["1"], 1)
	// // a["1"]  = append(a["1"], 2)
	// // a["2"]  = append(a["2"], 1)
	// // a["2"]  = append(a["2"], 2)

	// _, ok := a [ "1" ]
	// fmt.Print(ok)
	// txID := "3"
	// if a[txID] != nil{
	// 	fmt.Print("xx")
	// }
	// fmt.Print("yyy")

}