package blockchain

import (
	"scb/common"
)

type Block struct {
	header *BlockHeader
	hash   common.Hash
}
