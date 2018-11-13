package block

import "encoding/gob"
import "bytes"
import "log"
import "github.com/symphonyprotocol/sutil/elliptic"


// Transaction represents a Bitcoin transaction
type Transaction struct {
	Nonce   int64
	From    string
	To      string
	Amount  int64
	Signature []byte
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
//  Deserializes Transaction
func DeserializeTransction(d []byte) *Transaction {
	var transaction Transaction

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&transaction)
	if err != nil {
		log.Panic(err)
	}

	return &transaction
}

func (tx *Transaction) Sign(privKey *elliptic.PrivateKey){
	transbytes := tx.Serialize()
	sign_bytes, _ := elliptic.SignCompact(elliptic.S256(), privKey,  transbytes, true)
	tx.Signature = sign_bytes
}

func (tx *Transaction) Verify() bool{
	trans := NewTransaction(tx.Nonce, tx.Amount, tx.From, tx.To)
	transbytes := trans.Serialize()
	recover_pubkey, compressed, err := elliptic.RecoverCompact(elliptic.S256(), tx.Signature, transbytes)
	if err != nil || !compressed{
		return false
	}else{
		address := recover_pubkey.ToAddressCompressed()
		return address == tx.From
	}
}

func NewTransaction(nonce, amount int64, from, to string) *Transaction{
	trans := Transaction{
		Nonce : nonce,
		From : from,
		To : to,
		Amount: amount,
		Signature: []byte(""),
	}
	return &trans
}