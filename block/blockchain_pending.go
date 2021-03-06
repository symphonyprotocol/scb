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


func GetPendingBlock(blockHash []byte) *Block{
	var block *Block
	utils.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockPendingBucket))
		bytesBlock := b.Get(blockHash)
		if nil != bytesBlock{
			block = DeserializeBlock(bytesBlock)
		}
		return nil
	})
	return block
}
func GetSinglePendingBlock(blockHash []byte) *Block{
	var block *Block
	utils.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockPendingSingleBucket))
		bytesBlock := b.Get(blockHash)
		if nil != bytesBlock{
			block = DeserializeBlock(bytesBlock)
		}
		return nil
	})
	return block
}

func SavePendingBlock(block *Block){
	utils.Update(func(tx *bolt.Tx) error {
		bs := tx.Bucket([]byte(blockPendingBucket))
		bs.Put(block.Header.Hash, block.Serialize())
		return nil
	})
}
func SaveSinglePendingBlock(block *Block){
	utils.Update(func(tx *bolt.Tx) error {
		bs := tx.Bucket([]byte(blockPendingSingleBucket))
		bs.Put(block.Header.PrevBlockHash, block.Serialize())
		return nil
	})
}

func (bcp *BlockchainPendingPool) FindRootBlock(block *Block)(byte, *Block){
	var rootBlock *Block
	var pendingLength byte = 0
	fmt.Printf("--- Finding root block with block: %v\n", block.Header.HashString())

	utils.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockPendingBucket))
		var flagHash []byte = block.Header.PrevBlockHash
		fmt.Printf("--- flagHash init as: %v\n", ut.BytesToString(flagHash))
		var prevBlock *Block
		for {
			pendingLength += 1
			blockbytes := b.Get(flagHash)
			currentBlock := DeserializeBlock(blockbytes)
			if currentBlock != nil {
				fmt.Printf("--- hash of current block: %v\n", currentBlock.Header.HashString())
			} else {
				fmt.Println("--- current block is nil --- ")
			}
			fmt.Printf("--- hash of flaghash: %v\n", ut.BytesToString(flagHash))
			fmt.Printf("--- hash of bcp.Root: %v\n", ut.BytesToString(bcp.Root))
			if bytes.Compare(flagHash, bcp.Root) == 0{
				fmt.Printf("--- EQUALS: prevBlock: %v\n", prevBlock)
				if prevBlock == nil && currentBlock != nil {
					fmt.Printf("--- EQUALS: prevBlock is nil, use currentBlock\n")
					rootBlock = currentBlock
				}
				if prevBlock == nil && currentBlock == nil {
					fmt.Printf("--- EQUALS: the partial chain has only one block, it's itself\n")
					rootBlock = block
				}
				if prevBlock != nil {
					fmt.Printf("--- EQUALS: use prevBlock\n")
					rootBlock = prevBlock
				}
				break
			}
			prevBlock = currentBlock
			flagHash = currentBlock.Header.PrevBlockHash
		}
		return nil
	})
	if rootBlock == nil{
		pendingLength = 0
	}
	return pendingLength, rootBlock
}

func(bcp *BlockchainPendingPool) SavePendingBlockDetails(block *Block) (byte, []byte, *Block){
	var pendingLength byte
	var blocktailHash []byte
	var rootBlock *Block

	pendingLength, rootBlock = bcp.FindRootBlock(block)

	utils.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockPendingBucket))
		keyTail0 := append([]byte("lt"), rootBlock.Header.Hash...)
		keyTail := append(keyTail0, block.Header.Hash...)
		keyHeight := append([]byte("height"), block.Header.Hash...)

		prevTail := append(keyTail0, block.Header.PrevBlockHash...)
		b.Delete(prevTail)
		blocktailHash = block.Header.Hash
		b.Put(keyTail, block.Header.Hash)
		var heightval []byte
		heightval = append(heightval, byte(pendingLength))
		b.Put(keyHeight, heightval)
		return nil
	})

	return pendingLength, blocktailHash, rootBlock
}

func(bcp *BlockchainPendingPool) ConnectSinglePendingPool(block *Block, rootBlock *Block, pendlingLength byte, blocktailHash []byte)(byte, []byte){

	if pendlingLength < pendingBlockCnt{
		singleBlock := GetSinglePendingBlock(block.Header.Hash)
		
		for nil != singleBlock{
			pendlingLength += 1
			prevBlock := GetPendingBlock(singleBlock.Header.PrevBlockHash)
			//verify pow
			stateTree := bcp.DerivationPendingTree(prevBlock)
			if pow_res := singleBlock.VerifyPowV2(stateTree); !pow_res{
				log.Panic("block pow verify fail")
			}

			SavePendingBlock(singleBlock)

			// update tail pointer
			prev_tail_key := append([]byte("lt"), rootBlock.Header.Hash...)
			prev_tail_key = append(prev_tail_key, prevBlock.Header.Hash...)
			current_tail_key := append(prev_tail_key, singleBlock.Header.Hash...)
			blocktailHash = current_tail_key
			
			utils.Update(func(tx *bolt.Tx) error {
				b := tx.Bucket([]byte(blockPendingBucket))
				bs := tx.Bucket([]byte(blockPendingSingleBucket))
				b.Delete(prev_tail_key)
				b.Put(current_tail_key, singleBlock.Header.Hash)
				bs.Delete(singleBlock.Header.PrevBlockHash)
				return nil
			})
			singleBlock = GetSinglePendingBlock(singleBlock.Header.Hash)
		}
	}
	return pendlingLength, blocktailHash
}


func (bcp *BlockchainPendingPool) AcceptBlock(block *Block) *BlockChainPending{
	var blockchainPending *BlockChainPending

	verifyAcceptBlock(block)

	//can connect root hash
	if bytes.Compare(block.Header.PrevBlockHash, bcp.Root) == 0{
		fmt.Printf("can be connected to main chain, root tree hash: %v\n", ut.BytesToString(bcp.RootStateTree.Root.Hash))
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
		prevBlock := GetPendingBlock(block.Header.PrevBlockHash)
		fmt.Printf("can not be connected to main chain, root tree hash: %v\n", ut.BytesToString(bcp.RootStateTree.Root.Hash))
		if prevBlock == nil{
			SaveSinglePendingBlock(block)
		}else{
			fmt.Printf("prev block hash: %v\n", prevBlock.Header.HashString())

			stateTree := bcp.DerivationPendingTree(prevBlock)
			fmt.Printf("derived state tree hash: %v\n", ut.BytesToString(stateTree.Root.Hash))
			if pow_res := block.VerifyPowV2(stateTree); !pow_res{
				log.Panic("block pow verify fail")
			} else {
				fmt.Printf("block pow verified successfully", ut.BytesToString(stateTree.Root.Hash))
			}

			SavePendingBlock(block)

			pendingLength, blocktailHash, rootBlock := bcp.SavePendingBlockDetails(block)
			pendingLength, blocktailHash = bcp.ConnectSinglePendingPool(block, rootBlock, pendingLength, blocktailHash)

			if pendingLength >= pendingBlockCnt{
				blockchainPending = &BlockChainPending{
					Head: rootBlock.Header.Hash,
					Tail: &BlockChainPendingTail{
						ltail: blocktailHash,
						height: pendingLength,
					},
				}
			}
		}
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

func (bcp *BlockchainPendingPool) GetBlockPendingChains(block *Block) *BlockChainPending{

	for _, chain := range bcp.PendingChains{
		if bytes.Compare(chain.Tail.ltail, block.Header.Hash) == 0{
			return chain
		}
	}
	return nil
}

func (bcp *BlockchainPendingPool) DerivationPendingTree(block *Block) *MerkleTree{
	treebytes := bcp.RootStateTree.BreadthFirstSerialize()
	copyTree := DeserializeNodeFromData(treebytes)

	fmt.Printf("copied tree root hash: %v\n", ut.BytesToString(copyTree.Root.Hash))

	if block == nil{
		return copyTree
	}
	
	stateTree := copyTree
	accounts := stateTree.DeserializeAccount()
	for _, acc := range accounts {
		fmt.Printf("account : %v\n", acc)
	}
	chain := bcp.GetBlockPendingChains(block)
	// var chain *BlockChainPending = nil 
	if chain == nil{
		length, rootBlock := bcp.FindRootBlock(block)
		chain = &BlockChainPending{
			Head: rootBlock.Header.Hash,
			Tail: &BlockChainPendingTail{
				ltail: block.Header.Hash,
				height: length,
			},
		}
		
		fmt.Printf("constructed pending chain with head: %v and tail: %v and height: %v\n", rootBlock.Header.HashString(), block.Header.HashString(), length)
	} else {
		fmt.Printf("found pending chain with head: %v and tail: %v and height: %v\n", ut.BytesToString(chain.Head), ut.BytesToString(chain.Tail.ltail), chain.Tail.height)
	}

	blocks := chain.ConvertPendingBlockchain2Blocks()
	for idx := len(blocks)-1 ; idx >= 0 ; idx-- {
		block_ := blocks[idx]
		fmt.Printf("deriving tree at block: %v, prev: %v\n", block_.Header.HashString(), ut.BytesToString(block_.Header.PrevBlockHash))
		change_accounts, new_accounts := block_.PreProcessAccountBalance(accounts)
		stateTree, _ = stateTree.UpdateTree(change_accounts, new_accounts)
		fmt.Printf("stateTree updated to: %v\n", ut.BytesToString(stateTree.Root.Hash))
		accounts = stateTree.DeserializeAccount()
		for _, acc := range accounts {
			fmt.Printf("----account : %v\n", acc)
		}
		if bytes.Compare(block.Header.Hash, block_.Header.Hash) ==0 {
			break
		}
	}

	return stateTree
}

func ClearPendingPool(){
	utils.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket([]byte(blockPendingBucket))
		tx.DeleteBucket([]byte(blockPendingSingleBucket))
		tx.CreateBucketIfNotExists([]byte(blockPendingBucket))
		tx.CreateBucketIfNotExists([]byte(blockPendingSingleBucket))
		return nil
	})
}