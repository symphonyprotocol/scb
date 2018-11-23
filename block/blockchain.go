package block

import (
	"bytes"
	"github.com/boltdb/bolt"
	"os"
	"log"
	osuser "os/user"
	"fmt"
	"github.com/symphonyprotocol/sutil/elliptic"
	"github.com/symphonyprotocol/scb/utils"
)

const blocksBucket = "blocks"
const accountBucket = "account"
const transactionBucket = "transaction"
const gasBucket = "gas"
// 挖矿奖励金
const Subsidy = 100

var(
	CURRENT_USER, _ = osuser.Current()
	dbFile = CURRENT_USER.HomeDir + "/.blockchain.db"
)

type Blockchain struct {
	tip []byte
	// db  *bolt.DB
}

// func  (bc *Blockchain) GetDB() *bolt.DB{
// 	return bc.db
// }


// BlockchainIterator is used to iterate over blockchain blocks
type BlockchainIterator struct {
	currentHash []byte
	// db  *bolt.DB
}

// Iterator returns a BlockchainIterat
func (bc *Blockchain) Iterator() *BlockchainIterator {
	// bci := &BlockchainIterator{bc.tip, bc.db}
	_bc := LoadBlockchain()
	bc.tip = _bc.tip
	bci := &BlockchainIterator{bc.tip}
	return bci
}

// Next returns next block starting from the tip
func (i *BlockchainIterator) Next() *Block {
	var block *Block

	utils.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		encodedBlock := bucket.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)
		return nil
	})

	if block == nil {
		return nil
	}

	i.currentHash = block.Header.PrevBlockHash
	return block
}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}


// load Blockchain from db
func LoadBlockchain() *Blockchain {
	if dbExists() == false {
		fmt.Println("no existing blockchain, create one.")
		os.Exit(1)
	}

	var tip []byte

	utils.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip0 := b.Get([]byte("l"))

		tip = make([]byte, len(tip0))
		copy(tip, tip0)
		return nil
	})
	bc := Blockchain{tip}
	return &bc
}

// new empty blockchain, just the db initialized.
func CreateEmptyBlockchain() *Blockchain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		return LoadBlockchain()
	}

	var tip []byte
	
	utils.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(accountBucket))
		if err != nil {
			log.Panic(err)
		}

		_, err = tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		_, err2 := tx.CreateBucket([]byte(transactionBucket))
		if err2 != nil {
			log.Panic(err)
		}
		_, err2 = tx.CreateBucket([]byte(gasBucket))

		if err2 != nil {
			log.Panic(err)
		}

		return nil
	})

	bc := Blockchain{tip}
	return &bc
}

// new blockchain with genesis Block
func CreateBlockchain(wif string, callback func(*Blockchain)) {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	prikey, _ := elliptic.LoadWIF(wif)
	privateKey, publickey := elliptic.PrivKeyFromBytes(elliptic.S256(), prikey)
	address := publickey.ToAddressCompressed()
	account := InitAccount(address)

	var tip []byte

	trans := NewTransaction(account.Nonce, Subsidy, "", address, true)
	trans.Sign(privateKey)

	NewGenesisBlock(trans, address, func (genesis *Block) {
		genesis.Sign(privateKey)

		utils.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucket([]byte(accountBucket))
			if err != nil {
				log.Panic(err)
			}
	
			err = b.Put([]byte(address), account.Serialize())
			if err != nil {
				log.Panic(err)
			}
	
			b, err = tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				log.Panic(err)
			}
	
			_, err2 := tx.CreateBucket([]byte(transactionBucket))
			if err2 != nil {
				log.Panic(err)
			}

			_, err2 = tx.CreateBucket([]byte(gasBucket))
			if err2 != nil {
				log.Panic(err)
			}
	
			err = b.Put(genesis.Header.Hash, genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}
	
			err = b.Put([]byte("l"), genesis.Header.Hash)
			if err != nil {
				log.Panic(err)
			}
			tip = genesis.Header.Hash
	
			return nil
		})
		ChangeBalance(genesis.Header.Coinbase, Subsidy, true)
		bc := Blockchain{tip}
		if callback != nil {
			callback(&bc)
		}
	})
}

func(bc *Blockchain) SaveTransaction(trans *Transaction){
	utils.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(transactionBucket))
		err := b.Put(trans.ID, trans.Serialize())
		if err != nil {
			log.Panic(err)
		}

		return nil
	})
}

func(bc *Blockchain) FindUnpackTransactionById(id []byte) *Transaction{
	var transaction *Transaction

	utils.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(transactionBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			// fmt.Printf("key=%s, value=%s\n", k, v)
			trans := DeserializeTransction(v)
			if bytes.Compare(trans.ID, id) == 0 {
				transaction = trans
				break
			}
		}
		return nil
	})

	return transaction
}

func(bc *Blockchain) FindUnpackTransaction(address string) []* Transaction{
	var transactions []* Transaction

	utils.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(transactionBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			// fmt.Printf("key=%s, value=%s\n", k, v)
			trans := DeserializeTransction(v)
			if trans.From == address{
				transactions = append(transactions, trans)
			}
		}
		return nil
	})

	return transactions
}

func (bc *Blockchain) FindAllUnpackTransaction() map[string] []* Transaction {
	var trans_map map[string] []* Transaction
	trans_map = make(map[string] []* Transaction)

	utils.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(transactionBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			// fmt.Printf("key=%s, value=%s\n", k, v)
			trans := DeserializeTransction(v)
			trans_s, ok := trans_map [trans.From]
			if ok{
				trans_s = append(trans_s, trans)
				trans_map[trans.From] = trans_s
			}else{
				trans_map[trans.From] = []* Transaction{trans}
			}
		}
		return nil
	})
	return trans_map
}


func(bc *Blockchain) GetBlockHeight() int64{
	var lastBlock Block
	utils.View(func(tx *bolt.Tx) error{
		bucket := tx.Bucket([]byte(blocksBucket))
		blockhash := bucket.Get([]byte ("l"))
		blockdata := bucket.Get(blockhash)
		lastBlock = *DeserializeBlock(blockdata)
		return nil
	})
	return lastBlock.Header.Height
}

// MineBlock mines a new block with the provided transactions
func (bc *Blockchain) MineBlock(wif string, transactions []*Transaction, callback func(* Block)) *ProofOfWork {
	var lastHash []byte
	var lastHeight int64
	for _, tx := range transactions{
		if !tx.Verify(){
			log.Panic("ERROR: Invalid transaction")
		}
	}

	prikey, _ := elliptic.LoadWIF(wif)
	privateKey, publickey := elliptic.PrivKeyFromBytes(elliptic.S256(), prikey)
	address := publickey.ToAddressCompressed()
	// account := InitAccount(address)


	utils.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash0 := b.Get([]byte("l"))
		lastHash = make([]byte, len(lastHash0))
		copy(lastHash, lastHash0)
		blockbytes := b.Get(lastHash0)
		block := DeserializeBlock(blockbytes)
		lastHeight = block.Header.Height
		return nil
	})

	return NewBlock(transactions, lastHash, lastHeight + 1, address, func(block * Block){
		if nil != block{
			block.Sign(privateKey)
			callback(block)
		}
	})
}


func (bc *Blockchain) verifyNewBlock(block *Block){
	//1. verify block POW
	pow_res := block.VerifyPow()
	//2. verify block hash
	block_hash_res := bc.VerifyBlockHash(block)
	//3. verfiy transactions
	trans_res := false
	for _, trans := range block.Transactions{
		if trans.Verify(){
			trans_res = true
		}else{
			trans_res = false
			break
		}
	}
	if !pow_res{
		log.Panic("block pow verify fail")
	}
	if !block_hash_res{
		log.Panic("block hash fail")
	}
	if !trans_res{
		log.Panic("block transaction verify fail")
	}
}

func(bc *Blockchain) AcceptNewBlock(block *Block){
	var blockchain *Blockchain

	if len(bc.tip) != 0 {
		blockchain = LoadBlockchain()
	}else{
		blockchain = bc
	}

	if bc.HasBlock(block.Header.Hash){
		fmt.Errorf("block already exists")
		return 
	}

	blockchain.verifyNewBlock(block)
	blockchain.CombineBlock(block)

	// for _, trans := range block.Transactions{
	// 	if trans.From != ""{
	// 		ChangeBalance(trans.From, 0 - trans.Amount)
	// 	}
	// 	ChangeBalance(trans.To, trans.Amount)
	// }
	// *bc = *LoadBlockchain()
}

func (bc *Blockchain) CombineBlock(block *Block){
	utils.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(block.Header.Hash, block.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), block.Header.Hash)
		if err != nil {
			log.Panic(err)
		}
		bc.tip = block.Header.Hash
		fmt.Print(block.Header.Hash)
		return nil
	})
}

func (bc *Blockchain) VerifyBlockHash(b *Block) bool{
	var lastHash []byte
	var lastHeight int64


	if len(bc.tip) > 0 {
		utils.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(blocksBucket))
			lastHash0 := b.Get([]byte("l"))
			
			lastHash = make([]byte, len(lastHash0))
			copy(lastHash, lastHash0)

			blockbytes := b.Get(lastHash0)

			block := DeserializeBlock(blockbytes)
			lastHeight = block.Header.Height

			return nil
		})
	}else{
		lastHeight = -1
	}
	// verify prevhash
	
	hashCompRes := bytes.Compare(b.Header.PrevBlockHash, lastHash)

	hashVerify := b.VerifyHash()

	fmt.Printf("last height: %v, header height:%v", lastHeight, b.Header.Height)
	if hashCompRes == 0 && hashVerify && lastHeight + 1 == b.Header.Height{
		return true
	}
	return false
}

func (bc *Blockchain) HasBlock(hash []byte) bool {
	var exists bool = false
	utils.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		blockbytes := b.Get(hash)
		if nil != blockbytes{
			exists = true
		}
		return nil
	})
	return exists
}