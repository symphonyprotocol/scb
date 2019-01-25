package block

import "math"
import "math/big"
import "bytes"
// import "encoding/gob"
import "fmt"
import "time"
import "crypto/sha256"
import "github.com/symphonyprotocol/scb/utils"
import sutils "github.com/symphonyprotocol/sutil/utils"
import "github.com/symphonyprotocol/log"
import "github.com/symphonyprotocol/sutil/elliptic"
import "sort"
import _log "log"
import "strconv"
import "github.com/boltdb/bolt"

const targetBits = 6

var maxNonce = int64(math.MaxInt64)
var blockLogger = log.GetLogger("scb").SetLevel(log.TRACE)

type BlockHeader struct{
	Timestamp      int64
	Difficulty     int64
	PrevBlockHash  []byte
	Hash           []byte
	Nonce          int64
	Height		   int64
	Coinbase       string
	MerkleRootHash []byte
	MerkleRootAccountHash []byte
	// MerkleRootAccountGasHash []byte
  	Signature 	   []byte
}


type Block struct {
	Header BlockHeader
	Transactions  []*Transaction
	Content []byte
}

// ProofOfWork represents a proof-of-work
type ProofOfWork struct {
	block  *Block
	target *big.Int
	quitSign	chan struct{}
	isFinished	bool
	Cancelled	chan struct{}
	Done 		chan []byte
}


// Serializes the block
func (b *Block) Serialize() []byte {
	return sutils.ObjToBytes(b)
	// var result bytes.Buffer
	// encoder := gob.NewEncoder(&result)

	// err := encoder.Encode(b)
	// if err != nil {
	// 	blockLogger.Error("Failed to serialize the block: %v", err)
	// 	panic(err)
	// }
	// return result.Bytes()
}


// Deserializes a block
func DeserializeBlock(d []byte) *Block {
	var block Block = Block{}
	if err := sutils.BytesToObj(d, &block); err != nil {
		return nil
	} else {
		return &block
	}
	// decoder := gob.NewDecoder(bytes.NewReader(d))
	// err := decoder.Decode(&block)
	// if err != nil {
	// 	blockLogger.Trace("Failed to deserialize the block: %v", err)
	// 	return nil
	// }

	// return &block
}

// Hash transactions with merkle tree
func (b *Block) HashTransactions() []byte {
	// var transactions [][]byte
	var transactions []BlockContent
	for _, tx := range b.Transactions {
		//fmt.Printf("the tx inside when verifying pow: %v", tx)
		serializedTx := tx.Serialize()
		//fmt.Printf("serialized tx when verifying pow: %v", serializedTx)
		transactions = append(transactions, BlockContent{X : serializedTx})
	}

	mTree, err := NewTree(transactions)
	if err == nil{
		return mTree.MerkleRoot()
	}
	return nil
}

func(b *Block) GetAccountTree(preprocess bool) *MerkleTree{
	lastStateTree := GetLastMerkleTree()

	if preprocess{
		accounts := GetAllAccount()
		sort.Slice(accounts,func(i, j int) bool{
			return accounts[i].Index < accounts[j].Index
		})
		change_accounts, new_accounts := b.PreProcessAccountBalance(accounts)
		tree, err := lastStateTree.UpdateTree(change_accounts, new_accounts)
		if err == nil{
			return tree
		}
	}
	return lastStateTree
}


func (b *Block) Sign(privKey *elliptic.PrivateKey) *Block{
	blockbytes := b.Serialize()
	sign_bytes, _ := elliptic.SignCompact(elliptic.S256(), privKey, blockbytes, true)
	b.Header.Signature = sign_bytes
	b.Content = blockbytes
	return b
}

func (h BlockHeader) HashString() string {
	return sutils.BytesToString(h.Hash)
}

// NewProofOfWork builds and returns a ProofOfWork
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{b, target, make(chan struct{}), false, make(chan struct{}), make(chan []byte)}

	return pow
}

// Run performs a proof-of-work
func (pow *ProofOfWork) Run(callback func(int64, []byte))  {
	var hashInt big.Int
	var hash [32]byte
	var nonce int64 = 0

	fmt.Println("Mining a new block")
	go func() {
		QUIT:
		for nonce < maxNonce {
			select {
			case <- pow.quitSign:
				break QUIT
			default:
				data, err := pow.prepareData(nonce, true)

				if err != nil {
					continue
				}
		
				hash = sha256.Sum256(data)
				fmt.Printf("%d->%x\n", nonce, hash)
				hashInt.SetBytes(hash[:])
		
				if hashInt.Cmp(pow.target) == -1 {
					// found
					fmt.Printf("find:%x\n", hash)
					if callback != nil {
						pow.isFinished = true
						callback(nonce, hash[:])
					}
					break QUIT
				} else {
					nonce++
					time.Sleep(time.Millisecond * 5)
				}
			}
		}
	}()
}


// Run performs a proof-of-work
func (pow *ProofOfWork) Runv2(merkleRoot []byte,callback func(int64, []byte))  {
	var hashInt big.Int
	var hash [32]byte
	var nonce int64 = 0

	fmt.Println("Mining a new block")
	go func() {
		QUIT:
		for nonce < maxNonce {
			select {
			case <- pow.quitSign:
				break QUIT
				pow.Cancelled <- struct{}{}
			default:
				data, err := pow.prepareDatav2(merkleRoot, nonce)
				if err != nil {
					continue
				}
		
				hash = sha256.Sum256(data)
				// fmt.Printf("%d->%x\n", nonce, hash)
				hashInt.SetBytes(hash[:])
		
				if hashInt.Cmp(pow.target) == -1 {
					// found
					fmt.Printf("Mine: Found: %d->%x\n", nonce, hash)
					if callback != nil {
						pow.isFinished = true
						callback(nonce, hash[:])
					}
					pow.Done <- hash[:]
					break QUIT
				} else {
					nonce++
					time.Sleep(time.Millisecond * 5)
				}
			}
		}
	}()
}


func (pow *ProofOfWork) Stop() {
	pow.quitSign <- struct{}{}
}
func (pow *ProofOfWork) IsFinished() bool {
	return pow.isFinished
}

// NewBlock creates and returns Block
func NewBlock(transactions []*Transaction, prevBlockHash []byte, height int64, coinbase string,  callback func(*Block, *MerkleTree)) *ProofOfWork {
	// rootHash := 
	header := BlockHeader{
		Timestamp: time.Now().Unix(),
		PrevBlockHash: prevBlockHash,
		Hash: []byte{},
		Nonce: 0,
		Height: height,
		Coinbase : coinbase,
	}
	block := &Block{
		Header: header,
		Transactions: transactions,
	}
	blockRootHash := block.HashTransactions()
	accountTree := block.GetAccountTree(true)

	block.Header.MerkleRootHash = blockRootHash
	block.Header.MerkleRootAccountHash = accountTree.MerkleRoot()

	pow := NewProofOfWork(block)
	pow.Run(func (nonce int64, hash []byte) {
		block.Header.Hash = hash[:]
		block.Header.Nonce = nonce
		if callback != nil {
			callback(block, accountTree)
		}
	})
	return pow
}

func NewBlockV2(transactions []*Transaction, prevBlockHash []byte, height int64, coinbase string,  prevStateTree *MerkleTree, callback func(*Block, *MerkleTree)) *ProofOfWork {
	header := BlockHeader{
		Timestamp: time.Now().Unix(),
		PrevBlockHash: prevBlockHash,
		Hash: []byte{},
		Nonce: 0,
		Height: height,
		Coinbase : coinbase,
	}
	block := &Block{
		Header: header,
		Transactions: transactions,
	}

	blockRootHash := block.HashTransactions()
	block.Header.MerkleRootHash = blockRootHash
	
	var accountTree *MerkleTree
	//创世块
	if prevStateTree == nil && height == 0 && prevBlockHash == nil{
		accountTree = InitGenesisStateTree(coinbase)
	}else{
		accounts := prevStateTree.DeserializeAccount()
		change_accounts, new_accounts := block.PreProcessAccountBalance(accounts)
		accountTree, _ = prevStateTree.UpdateTree(change_accounts, new_accounts)
	}

	block.Header.MerkleRootAccountHash = accountTree.MerkleRoot()
	pow := NewProofOfWork(block)
	
	pow.Runv2(accountTree.MerkleRoot(), func (nonce int64, hash []byte) {
		block.Header.Hash = hash[:]
		block.Header.Nonce = nonce
		if callback != nil {
			callback(block, accountTree)
		}
	})
	return pow
}

func InitGenesisStateTree(coinbase string) *MerkleTree{
	var contents []BlockContent
	accountTo := InitAccount(coinbase, 1)
	accountTo.Balance += Subsidy
	accountTo.Nonce += 1
	contents = append(contents, BlockContent{
		X : accountTo.Serialize(),
		Dup: false,
	})
	accountTree, err := NewTree(contents)
	if err == nil{
		return accountTree
	}
	return nil
}

func GetLastMerkleTree() *MerkleTree{
	height := GetBlockHeight()
	if height < 0{
		return nil
	}

	height_str := strconv.FormatInt(height, 10)
	var tree  *MerkleTree = nil

	utils.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(stateTreeBucket))
		if bucket != nil{
			treebytes := bucket.Get([]byte (height_str))
			if len(treebytes) > 0{
				tree = DeserializeNodeFromData(treebytes)
			}
		}
		return nil
	})
	return tree
}
func GetMerkleTreeByHeight(height int64) *MerkleTree{
	height_str := strconv.FormatInt(height, 10)
	var tree  *MerkleTree = nil

	utils.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(stateTreeBucket))
		if bucket != nil{
			treebytes := bucket.Get([]byte (height_str))
			if len(treebytes) > 0{
				tree = DeserializeNodeFromData(treebytes)
			}
		}
		return nil
	})
	return tree
}

func (pow *ProofOfWork) prepareData(nonce int64, preprocess bool) (retBytes []byte, retErr interface{}) {
	// to avoid method call conflict with AcceptNewBlock.
	// TODO: need optimization
	defer func() {
		if err := recover(); err != nil {
			retBytes = nil
			retErr = err
		}
	}()
	data := bytes.Join(
		[][]byte{
			pow.block.Header.PrevBlockHash,
			pow.block.HashTransactions(),
			pow.block.GetAccountTree(preprocess).MerkleRoot(),
			utils.IntToHex(pow.block.Header.Timestamp),
			utils.IntToHex(int64(targetBits)),
			utils.IntToHex(nonce),
		},
		[]byte{},
	)

	return data, nil
}

func (pow *ProofOfWork) prepareDatav2(merkleRoot []byte , nonce int64) (retBytes []byte, retErr interface{}) {
	defer func() {
		if err := recover(); err != nil {
			retBytes = nil
			retErr = err
		}
	}()
	data := bytes.Join(
		[][]byte{
			pow.block.Header.PrevBlockHash,
			pow.block.HashTransactions(),
			merkleRoot,
			utils.IntToHex(pow.block.Header.Timestamp),
			utils.IntToHex(int64(targetBits)),
			utils.IntToHex(nonce),
		},
		[]byte{},
	)
	return data, nil
}


func (block *Block) prepareData(preprocess bool) []byte{
	data := bytes.Join(
		[][]byte{
			block.Header.PrevBlockHash,
			block.HashTransactions(),
			block.GetAccountTree(preprocess).MerkleRoot(),
			utils.IntToHex(block.Header.Timestamp),
			utils.IntToHex(int64(targetBits)),
			utils.IntToHex(block.Header.Nonce),
		},
		[]byte{},
	)
	return data
}

func (block *Block) prepareDataV2(merkleRoot []byte) []byte{
	data := bytes.Join(
		[][]byte{
			block.Header.PrevBlockHash,
			block.HashTransactions(),
			merkleRoot,
			utils.IntToHex(block.Header.Timestamp),
			// utils.IntToHex(block.Header.Difficulty),
			utils.IntToHex(int64(targetBits)),
			utils.IntToHex(block.Header.Nonce),
		},
		[]byte{},
	)
	return data
}


func (block *Block) VerifyPow(preprocess bool) bool{
	var hashInt big.Int
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	data := block.prepareData(preprocess)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	return hashInt.Cmp(target) == -1
}

func (block *Block) VerifyPowV2(prevStateTree *MerkleTree) bool{
	var hashInt big.Int
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	var stateTree *MerkleTree

	if block.Header.PrevBlockHash == nil{
		// stateTree = GetLastMerkleTree()
		stateTree = InitGenesisStateTree(block.Header.Coinbase)
	}else{
		var err error
		prevAccounts := prevStateTree.DeserializeAccount()
		change_accounts, new_accounts := block.PreProcessAccountBalance(prevAccounts)
		stateTree, err = prevStateTree.UpdateTree(change_accounts, new_accounts)
		if err != nil{
			_log.Panic(err)
		}
	}

	merkleRoot := stateTree.MerkleRoot()

	data := block.prepareDataV2(merkleRoot)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	return hashInt.Cmp(target) == -1
}

func(block *Block) VerifyMerkleHash() bool{
		var transactions []BlockContent
		for _, tx := range block.Transactions {
			transactions = append(transactions, BlockContent{X : tx.Serialize()})
		}
		mTree, err := NewTree(transactions)
		var calcHash []byte
		if err == nil{
			calcHash = mTree.MerkleRoot()
		}
		return bytes.Compare(calcHash, block.Header.MerkleRootHash) == 0
}

func (block *Block) VerifyHash() bool{
	data := block.prepareData(true)
	hash := sha256.Sum256(data)

	return bytes.Compare(hash[:], block.Header.Hash) == 0
}

// NewGenesisBlock creates and returns genesis Block
func NewGenesisBlock(trans *Transaction, coinbase string,  callback func(*Block, *MerkleTree)) {
	fmt.Println("New Genesis Block")
	NewBlockV2([]*Transaction{trans}, nil, 0, coinbase, nil, callback)
}

func (block *Block) VerifyCoinbase() bool{
	signature := block.Header.Signature

	blockbyts := block.Content
	recover_pubkey, compressed, err := elliptic.RecoverCompact(elliptic.S256(), signature, blockbyts)

	if err != nil || !compressed{
		return false
	}else{
		address := recover_pubkey.ToAddressCompressed()
		return address == block.Header.Coinbase
	}
 }

 func (block *Block) VerifyTransaction() bool{
	var res bool
	for _, trans := range block.Transactions{
		if trans.Verify(){
			res = true
		}else{
			res = false
			break
		}
	}
	return res
 }

 func (block *Block) PreProcessAccountBalance(accounts [] *Account) ([]*Account, []*Account){
		var changedAccounts []*Account
		var newAccounts [] *Account
		
		idx := int64(len(accounts)) + 1

		for _, v := range block.Transactions{
			account_from := FindAccount(accounts, v.From)
			account_to := FindAccount(accounts, v.To)
				if v.From == ""{
					// 创世交易
					if account_to == nil{
						account_to = InitAccount(v.To, idx)
						idx ++
						account_to.Nonce += 1
						accounts = append(accounts, account_to)
						newAccounts = append(newAccounts, account_to)
					}
				}else{
					if account_from == nil{
						_log.Panic(v.From, ": no this account")
					}
					if account_to == nil{
						account_to = InitAccount(v.To, idx)
						idx ++
						account_from.Nonce += 1
						account_to.Nonce += 1
						account_to.Balance += v.Amount
						account_from.Balance -= v.Amount
						accounts = append(accounts, account_to)
						newAccounts = append(newAccounts, account_to)

						if nil == FindAccount(changedAccounts, v.From){
							changedAccounts = append(changedAccounts, account_from)
						}
					}else{
						account_to.Balance += v.Amount
						account_from.Balance -= v.Amount
						account_from.Nonce += 1
						account_to.Nonce += 1
						if nil == FindAccount(changedAccounts, v.From){
							changedAccounts = append(changedAccounts, account_from)
						}
						if nil == FindAccount(changedAccounts, v.To) && nil == FindAccount(newAccounts, v.To){
							changedAccounts = append(changedAccounts, account_to)
						}
					}
					if account_from.Balance < 0{
						_log.Panic(v.From, ": has no enough amount to continue the transaction")
					}
				}
		}
		
		var coinbase_account *Account = nil
		coinbase_account = FindAccount(accounts, block.Header.Coinbase)
		if coinbase_account == nil {
			coinbase_account = InitAccount(block.Header.Coinbase, idx)
			newAccounts = append(newAccounts, coinbase_account)
		} else {
			changedAccounts = append(changedAccounts, coinbase_account)
		}
		// coinbase_account = FindAccount(changedAccounts, block.Header.Coinbase)
		// if coinbase_account == nil{
		// 	coinbase_account = FindAccount(newAccounts, block.Header.Coinbase)
		// 	if coinbase_account == nil{
		// 		coinbase_account = InitAccount(block.Header.Coinbase, idx)
		// 		newAccounts = append(newAccounts, coinbase_account)
		// 	}
		// }
		coinbase_account.Balance += Subsidy
		
		return changedAccounts, newAccounts
 }

 func (block *Block) SaveTransactions(){
	//save transaction
	utils.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(transactionMapBucket))
		for _, trans := range block.Transactions{
			err := b.Put(trans.ID, trans.Serialize())
			if err != nil {
				fmt.Printf("an error when save:%s", err.Error())
			}
		}
		return nil
	})
	blockLogger.Trace("txs saved")
 }

 func (block *Block) DeleteTransactions(){
	//delete packed transaction
	utils.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(transactionBucket))
		for _, trans := range block.Transactions{
			err := b.Delete(trans.ID)
			if err != nil {
				fmt.Printf("an error when delete:%s", err.Error())
			}
		}
		return nil
	})
 }

 func(block *Block) SaveAccounts(){
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
 }
