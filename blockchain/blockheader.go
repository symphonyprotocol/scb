package blockchain

import (
	"bytes"
	"scb/common"
)

const MaxBlockHeaderSize = 16 + (common.HashSize * 2)

type BlockHeader struct {
	PreBlock   common.Hash
	MerkleRoot common.Hash
	Height     int32
	TimeStamp  int64
	Difficulty uint32
	Nonce      uint32
	Version    int32
}

func (h *BlockHeader) BlockHash() common.Hash {
	buf := bytes.NewBuffer(make([]byte, 0, MaxBlockHeaderSize))
	return common.DoubleHash(buf)
}

func NewBlockHeader(version int32, preHash, merkRootHash *common.Hash, difficulty, nonce uint32) *BlockHeader {
	return &BlockHeader{
		PreBlock: *preHash,
		MerkleRoot: *merkRootHash,
		TimeStamp: time.Unix(time.Now().Unix(), 0)，
		Difficulty: difficulty,
		Nonce: Nonce,
	}
}


