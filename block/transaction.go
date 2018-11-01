package block
import "fmt"
import "crypto/sha256"
import "bytes"
import "encoding/gob"
import "log"
import "encoding/hex"
import "github.com/symphonyprotocol/sutil/elliptic"
import "math/big"

// 挖矿奖励金
const subsidy = 100

// Transaction represents a Bitcoin transaction
type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

// IsCoinbase checks whether the transaction is coinbase
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// Serialized Transaction
func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

// Hash returns the hash of the Transaction
func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}


func (tx *Transaction) Sign(privKey elliptic.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}
	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}
	txCopy := tx.TrimmedCopy()

	for inID, vin := range txCopy.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil

		
		// s, err := privKey.Sign(rand.Reader, &privKey, txCopy.ID)
		s, err := privKey.Sign(txCopy.ID)
		if err != nil {
			log.Panic(err)
		}
		signature := append(s.R.Bytes(), s.S.Bytes()...)

		tx.Vin[inID].Signature = signature
	}
}


// TrimmedCopy creates a trimmed copy of Transaction to be used in signing
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, nil})
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}

	return txCopy
}

// Verify verifies signatures of Transaction inputs
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			log.Panic("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	for inID, vin := range tx.Vin {
		prevTx := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil

		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}

		pubkey, err := elliptic.ParsePubKey(vin.PubKey, elliptic.S256())
		if err != nil{
			return false
		}
		x.SetBytes(pubkey.X.Bytes())
		y.SetBytes(pubkey.Y.Bytes())

		sig := elliptic.Signature{
			R: &r,
			S: &s,
		}
		return sig.Verify(txCopy.ID, pubkey)
	}

	return true
}


// 创建coinbase 交易
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TXInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTXOutput(subsidy, to)
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}}
	tx.ID = tx.Hash()

	return &tx
}

// SetID sets ID of a transaction using sha256 hash gob encode bytes
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

// // CanUnlockOutputWith checks whether the address initiated the transaction
// func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
// 	return in.ScriptSig == unlockingData
// }

// // CanBeUnlockedWith checks if the output can be unlocked with the provided data
// func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
// 	return out.ScriptPubKey == unlockingData
// }

// NewUTXOTransaction creates a new transaction
// func NewUTXOTransaction(from, to string, amount int, bc *Blockchain, privkey []byte) *Transaction {
// 	var inputs []TXInput
// 	var outputs []TXOutput

// 	privateKey, publicKey := elliptic.PrivKeyFromBytes(elliptic.S256(), privkey)
// 	pubkey := publicKey.SerializeCompressed()

// 	pubKeyHash := elliptic.HashPubKey(pubkey)
// 	acc, validOutputs := bc.FindSpendableOutputs(pubKeyHash, amount)
// 	if acc < amount {
// 		log.Panic("ERROR: Not enough funds")
// 	}

// 	// Build a list of inputs
// 	for txid, outs := range validOutputs {
// 		txID, err := hex.DecodeString(txid)
// 		if err != nil {
// 			log.Panic(err)
// 		}

// 		for _, out := range outs {
// 			input := TXInput{txID, out, nil, pubkey}
// 			inputs = append(inputs, input)
// 		}
// 	}
// 	// Build a list of outputs
// 	outputs = append(outputs, *NewTXOutput(amount, to))
// 	if acc > amount {
// 		outputs = append(outputs, *NewTXOutput(acc-amount, from)) // a change
// 	}

// 	tx := Transaction{nil, inputs, outputs}
// 	tx.ID = tx.Hash()
	
// 	bc.SignTransaction(&tx, *privateKey)
// 	return &tx
// }


// NewUTXOTransaction creates a new transaction
func NewUTXOTransaction(from, to string, amount int, utxoset *UTXOSet, privkey []byte) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	privateKey, publicKey := elliptic.PrivKeyFromBytes(elliptic.S256(), privkey)
	pubkey := publicKey.SerializeCompressed()

	pubKeyHash := elliptic.HashPubKey(pubkey)


	acc, validOutputs := utxoset.FindSpendableOutputs(pubKeyHash, amount)
	if acc < amount {
		log.Panic("ERROR: Not enough funds")
	}

	// Build a list of inputs
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := TXInput{txID, out, nil, pubkey}
			inputs = append(inputs, input)
		}
	}
	// Build a list of outputs
	outputs = append(outputs, *NewTXOutput(amount, to))
	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from)) // a change
	}

	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	
	utxoset.Blockchain.SignTransaction(&tx, *privateKey)
	return &tx
}