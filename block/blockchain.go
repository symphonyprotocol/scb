package block
import "github.com/boltdb/bolt"
import "os"
import "log"
import 	osuser "os/user"
import "fmt"
import "github.com/symphonyprotocol/sutil/elliptic"

const blocksBucket = "blocks"
const accountBucket = "account"
const packageBucket = "packages"
// 挖矿奖励金
const subsidy = 100

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

	if err != nil {
		log.Panic(err)
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

	err = db.Update(func(tx *bolt.Tx) error {
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


	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(accountBucket))
		if err != nil {
			log.Panic(err)
		}
		account := NewAccount(address)
		err = b.Put([]byte(address), account.Serialize())
		if err != nil {
			log.Panic(err)
		}
		trans := NewTransaction(account.Nonce, subsidy, "", address)
		trans.Sign(privateKey)
		NewGenesisBlock(trans, func (genesis *Block) {
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

			if err != nil {
				log.Panic(err)
			}
		
			bc := Blockchain{tip, db}
			if callback != nil {
				callback(&bc)
			}
		})
		return nil
	})
}
