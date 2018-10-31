package block

import "bytes"
import "github.com/symphonyprotocol/sutil/elliptic"

// TXInput represents a transaction input
type TXInput struct {
	// 之前交易的 ID
	Txid []byte
	// 之前交易输出索引
	Vout int
	// 交易签名
	Signature []byte
	// 公钥, 未hash
	PubKey    []byte
}

//UsesKey checks whether the address initiated the transaction
func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := elliptic.HashPubKey(in.PubKey)
	return bytes.Compare(lockingHash, pubKeyHash) == 0
}