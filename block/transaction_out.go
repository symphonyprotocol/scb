package block

import "bytes"
// import "github.com/symphonyprotocol/sutil/base58"
import "github.com/symphonyprotocol/sutil/elliptic"
import "log"
import "encoding/gob"

// TXOutput represents a transaction output
type TXOutput struct {
	// val of output
	Value      int
	//  hashed public key
	PubKeyHash []byte
}
// TXOutputs collects TXOutput
type TXOutputs struct {
	Outputs []TXOutput
}


// Lock signs the output
func (out *TXOutput) Lock(address string) {
	pubKeyHash, valid := elliptic.LoadAddress(address)
	if valid{
		out.PubKeyHash = pubKeyHash
	}else{
		log.Panic("address is not valid")
	}
}

// IsLockedWithKey checks if the output can be used by the owner of the pubkey
func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

// NewTXOutput create a new TXOutput
func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.Lock(address)
	return txo
}


// DeserializeOutputs deserializes TXOutputs
func DeserializeOutputs(data []byte) TXOutputs {
	var outputs TXOutputs

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs)
	if err != nil {
		log.Panic(err)
	}

	return outputs
}

// Serialize serializes TXOutputs
func (outs TXOutputs) Serialize() []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(outs)
	if err != nil {
		log.Panic(err)
	}
	return buff.Bytes()
}