package block

import (
	// "github.com/symphonyprotocol/scb/block"
	"bytes"
	"github.com/boltdb/bolt"
	"os"
	"log"
	osuser "os/user"
	"fmt"
	"github.com/symphonyprotocol/sutil/elliptic"
	"github.com/symphonyprotocol/scb/utils"
	// "encoding/binary"
	"strconv"
)

const blocksBucket = "blocks"
const blocksIndexBucket = "blocks_indexes"
const blocksIndex_heightPrefix = "h"
const accountBucket = "account"
const transactionBucket = "transaction_pool"
const transactionMapBucket = "transaction"
const stateTreeBucket = "statetree"
const accountCacheBucket = "account_cache"
// 挖矿奖励金
const Subsidy = 1000000000

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

func removeDB() error {
	err := os.Remove(dbFile)
	return err
}

func DeleteBlockchain() error {
	return removeDB()
}


// load Blockchain from db
func LoadBlockchain() *Blockchain {
	RegisterSCBTypes()
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

		_, err = tx.CreateBucket([]byte(blocksIndexBucket))
		if err != nil {
			log.Panic(err)
		}

		_, err = tx.CreateBucket([]byte(transactionMapBucket))
		if err != nil {
			log.Panic(err)
		}

		_, err2 := tx.CreateBucket([]byte(transactionBucket))
		if err2 != nil {
			log.Panic(err)
		}

		_, err2 = tx.CreateBucket([]byte(stateTreeBucket))
		if err2 != nil {
			log.Panic(err)
		}
		_, err2 = tx.CreateBucket([]byte(accountCacheBucket))
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
	fmt.Printf("address from wif %v\n", address)
	account := InitAccount(address, 0)

	// var tip []byte

	trans := NewTransaction(account.Nonce, Subsidy, "", address)
	trans = trans.Sign(privateKey)

	bc := CreateEmptyBlockchain()

	NewGenesisBlock(trans, address, func (genesis *Block, statetree *MerkleTree) {
		genesis = genesis.Sign(privateKey)
		bc.AcceptNewBlock(genesis, statetree)
		if callback != nil {
			callback(bc)
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
func(bc *Blockchain) DeleteTransaction(trans *Transaction){
	utils.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(transactionBucket))
		err := b.Delete(trans.ID)
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


func GetBlockHeight() int64{
	var lastBlock Block
	if dbExists(){
		utils.View(func(tx *bolt.Tx) error{
			bucket := tx.Bucket([]byte(blocksBucket))
			blockhash := bucket.Get([]byte ("l"))
			blockdata := bucket.Get(blockhash)
			if len(blockdata) > 0{
				lastBlock = *DeserializeBlock(blockdata)
			}
			return nil
		})
	}else{
		return 0
	}
	return lastBlock.Header.Height
}

func(bc *Blockchain) GetBlock(height int64) *Block{
	chain := LoadBlockchain()
	bci := chain.Iterator()

	for{
		b := bci.Next()
		if b == nil{
			return nil
		}
		if b.Header.PrevBlockHash == nil {
			break
		}
		if b.Header.Height == height{
			return b
		}
	}
	return nil
} 

func (bc *Blockchain) GetBlockByHash(hash []byte) *Block {
	var the_block *Block = nil
	if dbExists() {
		utils.View(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(blocksBucket))
			blockdata := bucket.Get(hash)
			if len(blockdata) > 0 {
				the_block = DeserializeBlock(blockdata)
			}
			return nil
		})
	}

	return the_block
}

func (bc *Blockchain) GetBlockByHeight(height int64) *Block {
	var blockHash []byte
	if dbExists() {
		utils.View(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(blocksIndexBucket))
			blockHash = bucket.Get(getHeightKeyForHash(height))
			return nil
		})
	}

	return bc.GetBlockByHash(blockHash)
}

// MineBlock mines a new block with the provided transactions
func (bc *Blockchain) MineBlock(wif string, transactions []*Transaction, callback func(* Block, *MerkleTree)) *ProofOfWork {
	var lastHash []byte
	var lastHeight int64
	for _, trans := range transactions{
		if !trans.Verify(){
			// utils.Update
			utils.Update(func(tx *bolt.Tx) error {
				bucket := tx.Bucket([]byte(transactionBucket))
				bucket.Delete(trans.ID)
				return nil
			})

			log.Panic("ERROR: Invalid transaction, delete it")
		}
	}

	prikey, _ := elliptic.LoadWIF(wif)
	privateKey, publickey := elliptic.PrivKeyFromBytes(elliptic.S256(), prikey)
	address := publickey.ToAddressCompressed()
	// account := InitAccount(address)
	fmt.Printf("address from wif %v\n", address)
	
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

	return NewBlock(transactions, lastHash, lastHeight + 1, address, func(block * Block, st *MerkleTree){
		if nil != block{
			block = block.Sign(privateKey)
			callback(block, st)
		}
	})
}

func (bc *Blockchain) verifyNewBlock(block *Block){
	blockLogger.Trace("verifying block: %v, %v", block.Header.Height, block.Header.HashString())
	//1. verify block POW
	fmt.Println("//1. verify block POW")
	if pow_res := block.VerifyPow(true); !pow_res{
		log.Panic("block pow verify fail")
	}
	//2. verfiy transactions
	blockLogger.Trace("//2. verfiy transactions")
	if trans_res := block.VerifyTransaction(); !trans_res{
		log.Panic("block transaction verify fail")
	}

	//3. verify block signature
	blockLogger.Trace("//3. verify block signature")
	if coinbase_res := block.VerifyCoinbase(); !coinbase_res{
		log.Panic("block signature verify fail")
	}

	//4. verify block merkle root hash
	blockLogger.Trace("//4. verify block merkle root hash")
	if merkleRes := block.VerifyMerkleHash(); merkleRes == false{
		log.Panic("merkle root hash verify fail")
	}
}


func(bc *Blockchain) AcceptNewBlock(block *Block, st *MerkleTree){
	blockLogger.Trace("accepting new block")
	var blockchain *Blockchain

	if len(bc.tip) != 0 {
		blockchain = LoadBlockchain()
	}else{
		blockchain = bc
	}
	blockLogger.Trace("blockchain loaded")
	//无冲突
	if existBlock := bc.GetBlockByHash(block.Header.Hash); existBlock == nil{
		blockchain.verifyNewBlock(block)
		blockLogger.Trace("block verified")
		
		//2. verify block hash
		if block_hash_res := bc.VerifyBlockHash(block);!block_hash_res{
			log.Panic("block hash fail")
		}
		blockLogger.Trace("block hash verified")

		blockchain.CombineBlock(block)
		blockLogger.Trace("block combined")
		postAcceptBlock(block, st)
		blockLogger.Trace("post accept block done")

	}else{
		fmt.Println("block already exists, check timestamp")
		if block.Header.Timestamp >= existBlock.Header.Timestamp{
			fmt.Errorf("block exist and this block is later then exist one")
		}else{
			// test, remember delete the comment
			blockchain.verifyNewBlock(block)
			RevertTo(block.Header.Height - 1)
			blockchain.CombineBlock(block)
			postAcceptBlock(block, st)
		}
	}
}

func postAcceptBlock(block *Block, st *MerkleTree){
	accounts := GetAllAccount()
	blockLogger.Trace("all account loaded")
	change_accounts, new_accounts := block.PreProcessAccountBalance(accounts)
	blockLogger.Trace("account balance preprocessed")
	
	utils.Update(func(tx *bolt.Tx) error {
			
		bucket := tx.Bucket([]byte(accountBucket))
		for _, v := range change_accounts{
			accountbytes := bucket.Get([]byte(v.Address))
			if accountbytes != nil{
				bucket.Delete([]byte(v.Address))
				bucket.Put([]byte(v.Address), v.Serialize())
			}
		}
		return nil
	})
	blockLogger.Trace("account changes saved")
	utils.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(accountBucket))
		for _, v := range new_accounts{
			bucket.Put([]byte(v.Address), v.Serialize())
		}
		return nil
	})
	blockLogger.Trace("new accounts saved")

	//save state tree
	height_str := strconv.FormatInt(block.Header.Height, 10)
	utils.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(stateTreeBucket))
		if bucket != nil{
			del_indx := block.Header.Height - 5
			bucket.Put([]byte (height_str), st.BreadthFirstSerialize())
			if del_indx > 0{
				bucket.Delete([]byte(strconv.FormatInt(del_indx, 10)))
			}
		}
		return nil
	})
	blockLogger.Trace("state tree saved")
	
	//save transaction
	utils.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(transactionMapBucket))
		for _, trans := range block.Transactions{
			err := b.Put(trans.ID, block.Header.Hash)
			if err != nil {
				log.Panic(err)
			}
		}
		return nil
	})
	blockLogger.Trace("txs saved")

	//delete packed transaction
	utils.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(transactionBucket))
		for _, trans := range block.Transactions{
			err := b.Delete(trans.ID)
			if err != nil {
				// log.Panic(err)
				fmt.Printf("an error when delete:%v", err.Error)
			}
		}
		return nil
	})
	blockLogger.Trace("packed txs deleted")


		// // reward miner
		// if block.Header.Height > 0{
		// 	// ChangeBalance(block.Header.Coinbase, Subsidy)
		// }
		
		// //save transaction
		// for _, trans := range block.Transactions{
		// 	utils.Update(func(tx *bolt.Tx) error {
		// 		b := tx.Bucket([]byte(transactionMapBucket))
		// 		err := b.Put(trans.ID, block.Header.Hash)
		// 		if err != nil {
		// 			log.Panic(err)
		// 		}
		// 		return nil
		// 	})
		// }

		// //delete packed transaction
		// for _, trans := range block.Transactions{
		// 	utils.Update(func(tx *bolt.Tx) error {
		// 		b := tx.Bucket([]byte(transactionBucket))
		// 		err := b.Delete(trans.ID)
		// 		if err != nil {
		// 			// log.Panic(err)
		// 			fmt.Printf("an error when delete:%v", err.Error)
		// 		}
		// 		return nil
		// 	})
		// }
		
		// //change balance
		// for _, v := range block.Transactions{
		// 	if v.From == ""{
		// 		// 创世交易
		// 		// ChangeBalance(v.To, v.Amount)
		// 		// NoncePlus(v.To)
		// 	}else{
		// 		// ChangeBalance(v.From, 0 - v.Amount)
		// 		// ChangeBalance(v.To, v.Amount)
		// 		// NoncePlus(v.From)
		// 	}
		// }

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

		b = tx.Bucket([]byte(blocksIndexBucket))
		err = b.Put(getHeightKeyForHash(block.Header.Height), block.Header.Hash)
		if err != nil {
			log.Panic(err)
		}
		// fmt.Print(block.Header.Hash)
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

	fmt.Printf("last height: %v, header height:%v\n", lastHeight, b.Header.Height)
	if hashCompRes == 0 && hashVerify && lastHeight + 1 == b.Header.Height{
		return true
	}
	return false
}

func (bc *Blockchain) HasBlock(hash []byte) *Block {
	var block *Block
	utils.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		blockbytes := b.Get(hash)
		if nil != blockbytes{
			block = DeserializeBlock(blockbytes)
		}
		return nil
	})
	return block
}

func getHeightKeyForHash(height int64) []byte {
	return []byte(fmt.Sprintf("%v%v", blocksIndex_heightPrefix, height))
}

// revert block chain to specific height
func RevertTo(Height int64){
	chain := LoadBlockchain()
	bci := chain.Iterator()

	for{
		b := bci.Next()
		if b.Header.PrevBlockHash == nil {
			break
		}		
		if b.Header.Height > Height{
			utils.Update(func(tx *bolt.Tx) error {
				bucket := tx.Bucket([]byte(blocksBucket))
				bucket.Delete(b.Header.Hash)
				bucket = tx.Bucket([]byte(blocksIndexBucket))
				bucket.Delete(getHeightKeyForHash(b.Header.Height))
				return nil
			})
			// remove reward miner
			// ChangeBalance(b.Header.Coinbase, 0 - Subsidy)
			//delete saved transaction
			for _, trans := range b.Transactions{
				utils.Update(func(tx *bolt.Tx) error {
					bucket := tx.Bucket([]byte(transactionMapBucket))
					err := bucket.Delete(trans.ID)
					if err != nil {
						log.Panic(err)
					}
					return nil
				})
			}

			//recovery deleted packed transaction
			for _, trans := range b.Transactions{
				utils.Update(func(tx *bolt.Tx) error {
					bucket := tx.Bucket([]byte(transactionBucket))
					err := bucket.Put(trans.ID, b.Header.Hash)
					if err != nil {
						// log.Panic(err)
						fmt.Printf("an error when delete:%v", err.Error)
					}
					return nil
				})
			}
			//recovery changed balance
			for _, v := range b.Transactions{
				if v.From == ""{
					// ChangeBalance(v.To, v.Amount)
				}else{
					// ChangeBalance(v.From, v.Amount)
					// ChangeBalance(v.To, 0 - v.Amount)
				}
			}
		}else if b.Header.Height == Height{
			utils.Update(func(tx *bolt.Tx) error {
				bucket := tx.Bucket([]byte(blocksBucket))
				bucket.Delete([]byte("l"))
				bucket.Put([]byte("l"), b.Header.Hash)
				return nil
			})	
		}
	}
}


func PrintChain() {
	bc := LoadBlockchain()
	bci := bc.Iterator()

	for {
		b := bci.Next()

		fmt.Printf("Previous hash: %x\n", b.Header.PrevBlockHash)
		fmt.Printf("Hash: %x\n", b.Header.Hash)
		fmt.Printf("CreateAt: %v\n", b.Header.Timestamp)
		fmt.Printf("Height:%d\n", b.Header.Height)
		fmt.Printf("Coinbase:%v\n", b.Header.Coinbase)
		fmt.Printf("merkle Root:%v\n", b.Header.MerkleRootHash)
		fmt.Printf("account state Root:%v\n", b.Header.MerkleRootAccountHash)
		// pow := NewProofOfWork(b)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(b.VerifyPow(false)))
		fmt.Printf("Signature Verify:%v \n", b.VerifyCoinbase())
		fmt.Println()

		if b.Header.PrevBlockHash == nil {
			break
		}
	}
}