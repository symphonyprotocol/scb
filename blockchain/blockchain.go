package blockchain

import (
	"scb/common"
	"scb/database"
	"sync"
	"log"
)

type BlockChain struct {
	db         database.Database
	timeSource MedianTimeSource
	chainLock  sync.RWMutex
	index      *blockIndex
}

func NewBlockChina(db symdb.Database) (*BlockChain, error) {

}

func (b *BlockChain) blockExists(hash *common.Hash) (bool, error) {
	if b.index.HaveBlock(hash) {
		return true, nil
	}
	var exists bool
	err := b.db.View(func(dbTx database.Tx) error {
		var err error
		exists, err = dbTx.HasBlock(hash)
		if err != nil || !exists {
			return err
		}
	})
}

func (b *BlockChain) HaveBlock() (bool, error) {
	exists, err := b.blockExists(hash)
	if err != nil {
		return false, err
	}
	return false, exists
}

func (b *BlockChain) AddBlock(block *Block) (bool, error) {
	b.chainLock.Lock()
	defer b.chainLock.Unlock()

	blockHash := block.Hash()
	
}


func (b *BlockChain) Exists() {

}
