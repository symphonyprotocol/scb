package block

import "math"
import "math/big"
import "bytes"
import "encoding/gob"
import "fmt"
import "time"
import "crypto/sha256"
import "github.com/symphonyprotocol/scb/utils"
import sutils "github.com/symphonyprotocol/sutil/utils"
import "github.com/symphonyprotocol/log"
import "github.com/symphonyprotocol/sutil/elliptic"

const targetBits = 8

var maxNonce = int64(math.MaxInt64)
var blockLogger = log.GetLogger("scb")

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
}

// ProofOfWork represents a proof-of-work
type ProofOfWork struct {
	block  *Block
	target *big.Int
	quitSign	chan struct{}
}


// Serializes the block
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		blockLogger.Error("Failed to serialize the block: %v", err)
		panic(err)
	}
	return result.Bytes()
}

// Deserializes a block
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		blockLogger.Error("Failed to deserialize the block: %v", err)
		return nil
	}

	return &block
}

// Hash transactions with merkle tree
func (b *Block) HashTransactions() []byte {
	// var transactions [][]byte
	var transactions []Content

	for _, tx := range b.Transactions {
		transactions = append(transactions, BlockContent{X : tx.Serialize()})
	}

	mTree, err := NewTree(transactions)
	if err == nil{
		return mTree.MerkleRoot()
	}
	return nil
}

func(b *Block) HashAccount() []byte{
	var contents [] Content
	accounts := GetAllAccount()
	for _, ac := range accounts {
		contents = append(contents, BlockContent{X : ac.Serialize()})
	}
	mTree, err := NewTree(contents)
	if err == nil{
		return mTree.MerkleRoot()
	}
	return nil
}

func (b *Block) Sign(privKey *elliptic.PrivateKey){
	blockbytes := b.Serialize()
	sign_bytes, _ := elliptic.SignCompact(elliptic.S256(), privKey,  blockbytes, true)
	b.Header.Signature = sign_bytes
}

func (h BlockHeader) HashString() string {
	return sutils.BytesToString(h.Hash)
}

// NewProofOfWork builds and returns a ProofOfWork
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{b, target, make(chan struct{})}

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
				data := pow.prepareData(nonce)
		
				hash = sha256.Sum256(data)
				fmt.Printf("%d->%x\n", nonce, hash)
				hashInt.SetBytes(hash[:])
		
				if hashInt.Cmp(pow.target) == -1 {
					// found
					fmt.Printf("find:%x\n", hash)
					if callback != nil {
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

func (pow *ProofOfWork) Stop() {
	pow.quitSign <- struct{}{}
}

// NewBlock creates and returns Block
func NewBlock(transactions []*Transaction, prevBlockHash []byte, height int64, coinbase string,  callback func(*Block, )) *ProofOfWork {
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
	block.Header.MerkleRootHash = blockRootHash

	pow := NewProofOfWork(block)
	pow.Run(func (nonce int64, hash []byte) {
		block.Header.Hash = hash[:]
		block.Header.Nonce = nonce
		if callback != nil {
			callback(block)
		}
	})
	return pow
}


func (pow *ProofOfWork) prepareData(nonce int64) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.Header.PrevBlockHash,
			pow.block.HashTransactions(),
			utils.IntToHex(pow.block.Header.Timestamp),
			utils.IntToHex(int64(targetBits)),
			utils.IntToHex(nonce),
		},
		[]byte{},
	)

	return data
}


// Validate validates block's PoW
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Header.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}

func (block *Block) prepareData() []byte{
	data := bytes.Join(
		[][]byte{
			block.Header.PrevBlockHash,
			block.HashTransactions(),
			utils.IntToHex(block.Header.Timestamp),
			utils.IntToHex(int64(targetBits)),
			utils.IntToHex(block.Header.Nonce),
		},
		[]byte{},
	)
	return data
}

func (block *Block) VerifyPow() bool{
	var hashInt big.Int
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	data := block.prepareData()
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])
	return hashInt.Cmp(target) == -1
}

func(block *Block) VerifyMerkleHash() bool{
		var transactions []Content
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
	data := block.prepareData()
	hash := sha256.Sum256(data)

	return bytes.Compare(hash[:], block.Header.Hash) == 0
}

// NewGenesisBlock creates and returns genesis Block
func NewGenesisBlock(trans *Transaction, coinbase string,  callback func(*Block)) {
	fmt.Println("New Genesis Block")
	NewBlock([]*Transaction{trans}, []byte{}, 0, coinbase, callback)
}

func (block *Block) VerifyCoinbase() bool{
	recover_pubkey, compressed, err := elliptic.RecoverCompact(elliptic.S256(), block.Header.Signature, block.Serialize())
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
