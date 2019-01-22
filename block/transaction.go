package block

import (
	"github.com/symphonyprotocol/sutil/utils"
	// "encoding/gob"
	// "bytes"
	"log"
	"github.com/symphonyprotocol/sutil/elliptic"
	"crypto/sha256"
	"github.com/boltdb/bolt"
	scbutils "github.com/symphonyprotocol/scb/utils"
	// "encoding/binary"
)


// Transaction represents a Bitcoin transaction
type Transaction struct {
	ID        []byte
	Nonce     int64
	From      string
	To        string
	Amount    int64
	Signature []byte
}

// Serialized Transaction
func (tx Transaction) Serialize() []byte {

	return utils.ObjToBytes(tx)
	// var encoded bytes.Buffer
	// enc := gob.NewEncoder(&encoded)
	// err := enc.Encode(tx)
	// if err != nil {
	// 	log.Panic(err)
	// }

	// return encoded.Bytes()
}
//  Deserializes Transaction
func DeserializeTransction(d []byte) *Transaction {
	var transaction Transaction = Transaction{}

	// decoder := gob.NewDecoder(bytes.NewReader(d))
	// err := decoder.Decode(&transaction)
	// if err != nil {
	// 	log.Panic(err)
	// }
	
	utils.BytesToObj(d, &transaction)
	return &transaction
}

func (tx *Transaction) IDString() string {
	return utils.BytesToString(tx.ID)
}

func (tx *Transaction) Sign(privKey *elliptic.PrivateKey) *Transaction{
	transbytes := tx.Serialize()
	sign_bytes, _ := elliptic.SignCompact(elliptic.S256(), privKey,  transbytes, true)
	tx.Signature = sign_bytes
	return tx
}

func (tx *Transaction) Verify() bool{
	trans := NewTransaction(tx.Nonce, tx.Amount, tx.From, tx.To)
	transbytes := trans.Serialize()
	recover_pubkey, compressed, err := elliptic.RecoverCompact(elliptic.S256(), tx.Signature, transbytes)
	if err != nil || !compressed{
		return false
	}else{
		address := recover_pubkey.ToAddressCompressed()
		if tx.From == "" {
			return address == tx.To
		}
		return address == tx.From
	}
}

func (tx *Transaction) SetID() {
	// var encoded bytes.Buffer
	// var hash [32]byte

	// enc := gob.NewEncoder(&encoded)
	// err := enc.Encode(tx)
	// if err != nil {
	// 	log.Panic(err)
	// }
	bytes := tx.Serialize()
	hash := sha256.Sum256(bytes)
	tx.ID = hash[:]
}


func NewTransaction(nonce, amount int64, from, to string) *Transaction{
	trans := Transaction{
		Nonce : nonce,
		From : from,
		To : to,
		Amount: amount,
		Signature: []byte(""),
	}
	trans.SetID()
	return &trans
}

func GetMaxUnpackNonce(transactions []* Transaction) int64{
	var nonce int64 = -1
	for _, trans := range transactions{
		if trans.Nonce > nonce{
			nonce = trans.Nonce
		}
	}
	return nonce
}

func SendTo(from, to string, amount int64, wif string) *Transaction {

	_, validFrom := elliptic.LoadAddress(from)
	_, validTo := elliptic.LoadAddress(to)
	prikey, _ := elliptic.LoadWIF(wif)
	private_key, _ := elliptic.PrivKeyFromBytes(elliptic.S256(), prikey)

	var trans * Transaction
	
	if !validFrom{
		log.Panic("ERROR: Sender address is not valid")
	}
	if !validTo{
		log.Panic("ERROR: Recipient address is not valid")
	}

	account := GetAccount(from)

	if account.Balance < amount{
		log.Panic("ERROR: No enougn amount")
	}

	
	bc := LoadBlockchain()

	unpacktransactions := bc.FindUnpackTransaction(from)
	if len(unpacktransactions) == 0{
		trans = NewTransaction(account.Nonce + 1, amount, from, to)
	}else{
		nonce := GetMaxUnpackNonce(unpacktransactions)
		trans = NewTransaction(nonce + 1, amount, from, to)
	}

	trans = trans.Sign(private_key)

	bc.SaveTransaction(trans)
	return trans
}

func Mine(wif string, callback func([]* Transaction)) *ProofOfWork {
	// bc := LoadBlockchain()

	bcp := LoadPendingPool()
	var transactions []* Transaction
	unpacktransactions := FindAllUnpackTransaction()


	if len(unpacktransactions) > 0{
		for key := range unpacktransactions{
			transactions = unpacktransactions[key]
			break
		}

	}else{
		log.Panic("no transaction can be mine")
	}

	provework := bcp.MineBlock(wif, transactions, func(block *Block, st *MerkleTree) {
		pendingblockChain := bcp.AcceptBlock(block)
		if pendingblockChain != nil{
			bc := LoadBlockchain()
			bc.AcceptNewPendingChain(pendingblockChain)
		}
		block.DeleteTransactions()
		callback(transactions)
	})

	return provework
}

func FindAllUnpackTransaction() map[string] []* Transaction {
	var trans_map map[string] []* Transaction
	trans_map = make(map[string] []* Transaction)

	scbutils.View(func(tx *bolt.Tx) error {
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