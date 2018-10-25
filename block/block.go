package block

import "bytes"
import "encoding/gob"
import "log"
import "crypto/sha256"
import "time"
import "math/big"
import "fmt"
import "math"
import  . "../utils"

const targetBits = 12
var maxNonce = math.MaxInt64

type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

// ProofOfWork represents a proof-of-work
type ProofOfWork struct {
	block  *Block
	target *big.Int
}


// Serialize serializes the block
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// DeserializeBlock deserializes a block
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}

// HashTransactions returns a hash of the transactions in the block
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

// NewProofOfWork builds and returns a ProofOfWork
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{b, target}

	return pow
}

// Run performs a proof-of-work
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Println("Mining a new block")
	for nonce < maxNonce {
		data := pow.prepareData(nonce)

		hash = sha256.Sum256(data)
		fmt.Printf("%s", ".")
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
			time.Sleep(time.Millisecond * 5)
		}
	}
	fmt.Printf("find:%x\n", hash)

	return nonce, hash[:]
}


// NewBlock creates and returns Block
func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), transactions, prevBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// NewGenesisBlock creates and returns genesis Block
func NewGenesisBlock(coinbase *Transaction) *Block {
	fmt.Println("New Genesis Block")
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.HashTransactions(),
			IntToHex(pow.block.Timestamp),
			IntToHex(int64(targetBits)),
			IntToHex(int64(nonce)),
		},
		[]byte{},
	)

	return data
}

// Validate validates block's PoW
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}
