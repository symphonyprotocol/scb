package block

import "math"
import "math/big"
import "bytes"
import "encoding/gob"
import "log"
import "fmt"
import "time"
import "crypto/sha256"
import "github.com/symphonyprotocol/scb/utils"

const targetBits = 12

var maxNonce = int64(math.MaxInt64)

type BlockHeader struct{
	Timestamp     int64
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int64
	Height		  int64
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
		log.Panic(err)
	}
	return result.Bytes()
}

// Deserializes a block
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}

// Hash transactions with merkle tree
func (b *Block) HashTransactions() []byte {
	var transactions [][]byte

	for _, tx := range b.Transactions {
		transactions = append(transactions, tx.Serialize())
	}
	mTree := NewMerkleTree(transactions)

	return mTree.RootNode.Data
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
		for nonce < maxNonce {
			QUIT:
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
func NewBlock(transactions []*Transaction, prevBlockHash []byte, height int64, callback func(*Block)) {
	header := BlockHeader{
		Timestamp: time.Now().Unix(),
		PrevBlockHash: prevBlockHash,
		Hash: []byte{},
		Nonce: 0,
		Height: height,
	}
	block := &Block{
		Header: header,
		Transactions: transactions,
	}
	pow := NewProofOfWork(block)
	pow.Run(func (nonce int64, hash []byte) {
		block.Header.Hash = hash[:]
		block.Header.Nonce = nonce
		if callback != nil {
			callback(block)
		}
	})
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

// NewGenesisBlock creates and returns genesis Block
func NewGenesisBlock(trans *Transaction, callback func(*Block)) {
	fmt.Println("New Genesis Block")
	NewBlock([]*Transaction{trans}, []byte{}, 0, callback)
}