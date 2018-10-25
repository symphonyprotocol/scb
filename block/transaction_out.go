package block

import "bytes"
import "github.com/symphonyprotocol/sutil/base58"

// TXOutput represents a transaction output
type TXOutput struct {
	// val of output
	Value      int
	//  hashed public key
	PubKeyHash []byte
}

// Lock signs the output
func (out *TXOutput) Lock(address string) {
	pubKeyHash, _ := base58.B58decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
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
