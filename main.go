package main

// import "fmt"
// import "strconv"
// // import . "./block"


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
	// flag := make(chan struct{})

	//  block.CreateBlockchain("1T3r9yFFM6St9wGSp7zMYP24G6pUYnL7y", "L5fR7FRHnZGL3DjsrhN8CvBYHpywL8LjxA2rjzbL7qvFqjgbNVQ5", func(bc *block.Blockchain){
	// 	db := bc.GetDB()
	// 	db.Close()
	// 	block.ChangeBalance("1T3r9yFFM6St9wGSp7zMYP24G6pUYnL7y", block.Subsidy)
	// 	flag <- struct{}{}
	// })
	// <-flag

	// utxoset := block.UTXOSet{
	// 	Blockchain: bc,
	// }
	// utxoset.Reindex()


	// 2. get balance.

	// keyHashed, valid := elliptic.LoadAddress("1T3r9yFFM6St9wGSp7zMYP24G6pUYnL7y")

	// if !valid {
	// 	log.Panic("ERROR: Address is not valid")
	// }

	// bc := block.LoadBlockchain()
	// utxoset := block.UTXOSet{
	// 	Blockchain : bc,
	// }
	// db := bc.GetDB()
	// defer db.Close()


	// balance := 0
	// utxos := utxoset.FindUTXO(keyHashed)
	// for _, out := range utxos{
	// 	balance += out.Value
	// }
	


	// 3. send coin
	
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
	// utxoset := block.UTXOSet{
	// 	Blockchain: bc,
	// }

	// db := bc.GetDB()
	// defer db.Close()

	// tx := block.NewUTXOTransaction(from, to, amount, &utxoset, prikey)
	// fmt.Print(tx)
	// cbtx := block.NewCoinbaseTX(from, "")
	// txs := []* block.Transaction{cbtx, tx}

	// newblock := bc.MineBlock(txs)
	// utxoset.Update(newblock)
	// // bc.SaveUTXOTransaction(tx)

	// fmt.Println("Success!")


	// 3. mine
	// _, validaddress := elliptic.LoadAddress("1T3r9yFFM6St9wGSp7zMYP24G6pUYnL7y")

	// if !validaddress{
	// 	log.Panic("ERROR: Sender address is not valid")
	// }

	// bc2 := block.LoadBlockchain()
	
	// db2 := bc2.GetDB()
	// defer db2.Close()

	// cbtx := block.NewCoinbaseTX("1T3r9yFFM6St9wGSp7zMYP24G6pUYnL7y", "")
	// tx2 := bc2.LoadUTXOTransaction()
	// txs := []* block.Transaction{cbtx, tx2}
	// utxoset2 := block.UTXOSet{
	// 	Blockchain: bc2,
	// }

	// newblock := bc2.MineBlock(txs)
	// utxoset2.Update(newblock)




	// fmt.Printf("Balance of '%s': %d\n", address, balance)

	
	// 4.print 
	// bc2 := block.LoadBlockchain()
	// // db2 := bc2.GetDB()
	// // defer db2.Close()

	// bci := bc2.Iterator()

	// for {
	// 	b := bci.Next()

	// 	fmt.Printf("Prev. hash: %x\n", b.Header.PrevBlockHash)
	// 	fmt.Printf("Hash: %x\n", b.Header.Hash)
	// 	pow := block.NewProofOfWork(b)
	// 	fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
	// 	fmt.Println()

	// 	if len(b.Header.PrevBlockHash) == 0 {
	// 		break
	// 	}
	// }



	// 5. get ballance new
	// block.GetBalance("1T3r9yFFM6St9wGSp7zMYP24G6pUYnL7y")


	// 6. send
	//block.SendTo("1T3r9yFFM6St9wGSp7zMYP24G6pUYnL7y", "189wh8VjXLmKSZhnP9DQwcVKfvNemQSmBp", 1, "L5fR7FRHnZGL3DjsrhN8CvBYHpywL8LjxA2rjzbL7qvFqjgbNVQ5")
	
	//7. mine
	// block.Mine("1T3r9yFFM6St9wGSp7zMYP24G6pUYnL7y")
	// block.ChangeBalance("1T3r9yFFM6St9wGSp7zMYP24G6pUYnL7y", block.Subsidy)

	// block.GetBalance("1T3r9yFFM6St9wGSp7zMYP24G6pUYnL7y")
	

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