package block

import (
	"github.com/boltdb/bolt"
	"os"
	"log"
	osuser "os/user"
	"fmt"
	"github.com/symphonyprotocol/sutil/elliptic"
)

const blocksBucket = "blocks"
const accountBucket = "account"
const packageBucket = "packages"
// 挖矿奖励金
const Subsidy = 100

var(
	CURRENT_USER, _ = osuser.Current()
	dbFile = CURRENT_USER.HomeDir + "/.blockchain.db"
)

type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

func  (bc *Blockchain) GetDB() *bolt.DB{
	return bc.db
}


// BlockchainIterator is used to iterate over blockchain blocks
type BlockchainIterator struct {
	currentHash []byte
	db  *bolt.DB
}

// Iterator returns a BlockchainIterat
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}
	return bci
}

// Next returns next block starting from the tip
func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))
		encodedBlock := bucket.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)

		return nil
	})

	if err != nil || block == nil {
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
	db, err := bolt.Open(dbFile, 0600, nil)
	// defer db.Close()
	if err != nil {
		log.Panic(err)
	}

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}

// new empty blockchain, just the db initialized.
func CreateEmptyBlockchain() *Blockchain {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		return LoadBlockchain()
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	// defer db.Close()
	if err != nil {
		log.Panic(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte(accountBucket))
		if err != nil {
			log.Panic(err)
		}

		_, err = tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		_, err2 := tx.CreateBucket([]byte(packageBucket))
		if err2 != nil {
			log.Panic(err)
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}
	return &bc
}

// new blockchain with genesis Block
func CreateBlockchain(address, wif string, callback func(*Blockchain)) {
	if dbExists() {
		fmt.Println("Blockchain already exists.")
		os.Exit(1)
	}

	prikey, _ := elliptic.LoadWIF(wif)
	privateKey, _ := elliptic.PrivKeyFromBytes(elliptic.S256(), prikey)
	
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	// defer db.Close()
	if err != nil {
		log.Panic(err)
	}

	account := InitAccount(address)
	trans := NewTransaction(account.Nonce, Subsidy, "", address)
	trans.Sign(privateKey)

	NewGenesisBlock(trans, func (genesis *Block) {
		err = db.Update(func(tx *bolt.Tx) error {
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
	
			_, err2 := tx.CreateBucket([]byte(packageBucket))
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
		if err != nil {
			log.Panic(err)
		}
		bc := Blockchain{tip, db}
		if callback != nil {
			callback(&bc)
		}
	})
}

func(bc *Blockchain) SaveTransaction(trans *Transaction){
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(packageBucket))
		err := b.Put(trans.ID, trans.Serialize())
		if err != nil {
			log.Panic(err)
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}

func(bc *Blockchain) FindUnpackTransaction(address string) []* Transaction{
	var transactions []* Transaction

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(packageBucket))
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

	if err != nil {
		log.Panic(err)
	}

	return transactions
}

func (bc *Blockchain) FindAllUnpackTransaction() map[string] []* Transaction {
	var trans_map map[string] []* Transaction
	trans_map = make(map[string] []* Transaction)

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(packageBucket))
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
	if err != nil {
		log.Panic(err)
	}
	return trans_map
}


func(bc *Blockchain) GetBlockHeight() int64{
	var lastBlock Block
	err := bc.db.View(func(tx *bolt.Tx) error{
		bucket := tx.Bucket([]byte(blocksBucket))
		blockhash := bucket.Get([]byte ("l"))
		blockdata := bucket.Get(blockhash)
		lastBlock = *DeserializeBlock(blockdata)
		return nil
	})
	if err != nil{
		log.Panic(err)
	}
	return lastBlock.Header.Height
}

// MineBlock mines a new block with the provided transactions
func (bc *Blockchain) MineBlock(transactions []*Transaction, callback func(* Block)) {
	var lastHash []byte
	var lastHeight int64
	for _, tx := range transactions{
		if !tx.Verify(){
			log.Panic("ERROR: Invalid transaction")
		}
	}

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		
		blockbytes := b.Get(lastHash)
		block := DeserializeBlock(blockbytes)
		lastHeight = block.Header.Height
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	NewBlock(transactions, lastHash, lastHeight + 1, func(block * Block){
		if nil != block{
			callback(block)
		}
	})
}
