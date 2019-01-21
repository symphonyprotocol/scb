package block
import (
	"bytes"
	"github.com/symphonyprotocol/scb/utils"
	ut "github.com/symphonyprotocol/sutil/utils"
	"github.com/boltdb/bolt"
	"github.com/symphonyprotocol/sutil/elliptic"
	"log"
	"fmt"
)


type BlockChainPending struct{
	Head []byte
	Tail *BlockChainPendingTail
}

type BlockChainPendingTail struct{
	ltail []byte
	height byte
}


type BlockchainPendingPool struct{
	Root []byte
	RootHeight int64
	RootStateTree *MerkleTree
	PendingChains [] *BlockChainPending
}


func (bcp *BlockchainPendingPool) AcceptBlock(block *Block) *BlockChainPending{
	var pendlingLength byte
	var blockchainPending *BlockChainPending
	var blocktailHash []byte

	verifyAcceptBlock(block)

	//can connect root hash
	if bytes.Compare(block.Header.PrevBlockHash, bcp.Root) == 0{

		if pow_res := block.VerifyPowV2(bcp.RootStateTree); !pow_res{
			log.Panic("block pow verify fail")
		}
		
		var height []byte 
		utils.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(blockPendingBucket))
			val := b.Get(block.Header.Hash)

			if nil == val{
				var keyHead []byte
				var keyTail []byte

				keyHead = append([]byte("lh"), block.Header.Hash...)
				keyTail = append([]byte("lt"), block.Header.Hash...)
				keyTail = append(keyTail, block.Header.Hash...)
				keyHeight := append([]byte("height"), block.Header.Hash...)

				b.Put(block.Header.Hash, block.Serialize())
				b.Put(keyHead, block.Header.Hash)
				b.Put(keyTail, block.Header.Hash)
				height = append(height, byte(1))
				b.Put(keyHeight, height)
			}
			return nil
		})
	}else{
		utils.Update(func(tx *bolt.Tx) error {

			b := tx.Bucket([]byte(blockPendingBucket))
			bs := tx.Bucket([]byte(blockPendingSingleBucket))

			bytesPrevious := b.Get(block.Header.PrevBlockHash)
			if nil == bytesPrevious{
				//single block, put into single block pool
				bs.Put(block.Header.PrevBlockHash, block.Serialize())
			}else{
				//can connect with previous pending block
				block_serialize := block.Serialize()

				//verify pow
				blockPrevious := DeserializeBlock(bytesPrevious)
				stateTree := bcp.DerivationPendingTree(blockPrevious)
				if pow_res := block.VerifyPowV2(stateTree); !pow_res{
					log.Panic("block pow verify fail")
				}

				b.Put(block.Header.Hash, block_serialize)

				var flagHash []byte = block.Header.Hash
				var rootBlock *Block

				for bytes.Compare(flagHash, bcp.Root) != 0{
					pendlingLength += 1
					blockbytes := b.Get(flagHash)
					prevBlock := DeserializeBlock(blockbytes)
					flagHash = prevBlock.Header.PrevBlockHash
					if bytes.Compare(flagHash, bcp.Root) == 0{
						rootBlock = prevBlock
					}
				}


				keyTail0 := append([]byte("lt"), rootBlock.Header.Hash...)
				keyTail := append(keyTail0, block.Header.Hash...)
				keyHeight := append([]byte("height"), block.Header.Hash...)

				prevTail := append(keyTail0, block.Header.PrevBlockHash...)
				b.Delete(prevTail)
				blocktailHash = block.Header.Hash
				b.Put(keyTail, block.Header.Hash)
				var heightval []byte
				heightval = append(heightval, byte(pendlingLength))
				b.Put(keyHeight, heightval)


				// check if can connect from single block pending pool 
				if pendlingLength < pendingBlockCnt{
					singleBytes:= bs.Get(block.Header.Hash)

					for singleBytes != nil{
						pendlingLength += 1
						singleBlock := DeserializeBlock(singleBytes)
						privBytes := b.Get(singleBlock.Header.PrevBlockHash)
						prevBlock := DeserializeBlock(privBytes)

						//verify pow
						stateTree := bcp.DerivationPendingTree(prevBlock)
						if pow_res := singleBlock.VerifyPowV2(stateTree); !pow_res{
							log.Panic("block pow verify fail")
						}

						// append to pending blockchain
						b.Put(singleBlock.Header.Hash, singleBlock.Serialize())

						// update tail pointer
						prev_tail_key := append([]byte("lt"), rootBlock.Header.Hash...)
						prev_tail_key = append(prev_tail_key, prevBlock.Header.Hash...)
						current_tail_key := append(prev_tail_key, singleBlock.Header.Hash...)

						blocktailHash = current_tail_key

						b.Delete(prev_tail_key)
						b.Put(current_tail_key, singleBlock.Header.Hash)

						bs.Delete(singleBlock.Header.PrevBlockHash)
						singleBytes = bs.Get(singleBlock.Header.Hash)
					}
				}
				if pendlingLength >= pendingBlockCnt{
					blockchainPending = &BlockChainPending{
						Head: rootBlock.Header.Hash,
						Tail: &BlockChainPendingTail{
							ltail: blocktailHash,
						},
					}
				}
			}
			return nil
		})
	}
	return blockchainPending
}

func verifyAcceptBlock(block *Block){
	//verfiy transactions
	fmt.Println(". verfiy transactions")
	if trans_res := block.VerifyTransaction(); !trans_res{
		log.Panic("block transaction verify fail")
	}

	//verify block signature
	fmt.Println(". verify block signature")
	if coinbase_res := block.VerifyCoinbase(); !coinbase_res{
		log.Panic("block signature verify fail")
	}

	//verify block merkle root hash
	fmt.Println(". verify block merkle root hash")
	if merkleRes := block.VerifyMerkleHash(); merkleRes == false{
		log.Panic("merkle root hash verify fail")
	}
}

func(bc *BlockChainPending) ConvertPendingBlockchain2Blocks() [] *Block{
	var blocks [] *Block
	utils.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockPendingBucket))

		blockbytes := bucket.Get(bc.Tail.ltail)
		for {
			block := DeserializeBlock(blockbytes)
			blocks = append(blocks, block)
			if bytes.Compare(block.Header.Hash, bc.Head) == 0{
				break
			}
			blockhash := block.Header.PrevBlockHash
			blockbytes = bucket.Get(blockhash)
		}
		return nil
	})

	return blocks
}

func(bc *BlockChainPending) GetLastBlock() *Block{
	var block *Block
	utils.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockPendingBucket))
		blockbytes := bucket.Get(bc.Tail.ltail)
		block = DeserializeBlock(blockbytes)
		return nil
	})
	return block
}

func LoadPendingPool() *BlockchainPendingPool{
	lastBlock := GetLastBlock()
	lastStateTree := GetLastMerkleTree()


	var chains [] *BlockChainPending

	var mhead []string

	utils.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(blockPendingBucket))
		ch := bucket.Cursor()
		prefixH := []byte("lh")

		for k, _ := ch.Seek(prefixH); k != nil && bytes.HasPrefix(k, prefixH); k, _ = ch.Next() {
			key0 := k[2:]
			key := ut.BytesToString(key0)
			mhead = append(mhead, key)
		}
		return nil
	})

	for _, head := range mhead{
		utils.View(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(blockPendingBucket))
			ch := bucket.Cursor()

			headbytes := ut.StringToBytes(head)

			prefixT := append([]byte("lt"), headbytes...)

			for k, _ := ch.Seek(prefixT); k != nil && bytes.HasPrefix(k, prefixT); k, _ = ch.Next() {
				tailbytes := k[34:]
				
				tailbytes0 := make([]byte, len(tailbytes))
				copy(tailbytes0, tailbytes)

				keyHeight := append([]byte("height"), tailbytes...)
				bytesHeight := bucket.Get(keyHeight)
				height := bytesHeight[0]
				chains = append(chains, &BlockChainPending{
					Head: headbytes,
					Tail: &BlockChainPendingTail{
						height: height,
						ltail: tailbytes0,
					},
				})
			}
			return nil
		})
	}
	

	return &BlockchainPendingPool{
		Root: lastBlock.Header.Hash,
		RootHeight: lastBlock.Header.Height,
		PendingChains: chains,
		RootStateTree: lastStateTree,
	}
}

func (bcp *BlockchainPendingPool) MineBlock(wif string, transactions []*Transaction,  callback func(* Block, *MerkleTree)) *ProofOfWork {
	var lastBlock *Block

	longchain := bcp.GetLongChain()
	if longchain != nil{
		lastBlock = longchain.GetLastBlock()
	}
	stateTree := bcp.DerivationPendingTree(lastBlock)

	prikey, _ := elliptic.LoadWIF(wif)
	privateKey, publickey := elliptic.PrivKeyFromBytes(elliptic.S256(), prikey)
	address := publickey.ToAddressCompressed()

	var lastHash []byte
	var lastHeight int64

	if lastBlock == nil{
		lastHash = bcp.Root
		lastHeight = bcp.RootHeight + 1 
	}else{
		lastHash = lastBlock.Header.Hash
		lastHeight = bcp.RootHeight + int64(longchain.Tail.height) +1
	}

	return NewBlockV2(transactions, lastHash, lastHeight, address, stateTree, func(block * Block, st *MerkleTree){
		if nil != block{
			block = block.Sign(privateKey)
			callback(block, st)
		}
	})
}

func (bcp *BlockchainPendingPool) GetLongChain() *BlockChainPending{
	var longestChain *BlockChainPending

	if len(bcp.PendingChains) > 0{
		longestChain = bcp.PendingChains[0]
	}

	for idx, chain := range bcp.PendingChains{
		if idx == 0{
			continue
		}
		if chain.Tail.height > longestChain.Tail.height{
			longestChain = chain
		}
	}
	return longestChain
}

func (bcp *BlockchainPendingPool) DerivationPendingTree(block *Block) *MerkleTree{
	treebytes := bcp.RootStateTree.BreadthFirstSerialize()
	copyTree := DeserializeNodeFromData(treebytes)

	if block == nil{
		return copyTree
	}
	
	stateTree := copyTree
	accounts := stateTree.DeserializeAccount()
	longChain := bcp.GetLongChain()

	if longChain != nil{
		blocks := longChain.ConvertPendingBlockchain2Blocks()

		for idx := len(blocks)-1 ; idx >= 0 ; idx-- {
			block_ := blocks[idx]
			change_accounts, new_accounts := block_.PreProcessAccountBalance(accounts)
			stateTree, _ = stateTree.UpdateTree(change_accounts, new_accounts)
			accounts = stateTree.DeserializeAccount()
			if bytes.Compare(block.Header.Hash, block_.Header.Hash) ==0 {
				break
			}
		}
	}
	return stateTree
}