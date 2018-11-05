package block
import "crypto/sha256"

type MerkleTree struct{
	RootNode *MerkleTreeNode
}

type MerkleTreeNode struct{
	Left *MerkleTreeNode
	Right *MerkleTreeNode
	Data []byte
}

func NewMerkleTreeNode(left, right *MerkleTreeNode, data []byte) *MerkleTreeNode{
	node := MerkleTreeNode{}

	if left == nil && right == nil{
		hash := sha256.Sum256(data)
		node.Data = hash[:]
	}else{
		newData := append(left.Data, right.Data...)
		hash := sha256.Sum256(newData)
		node.Data = hash[:]
	}
	node.Left = left
	node.Right = right

	return &node
}

func NewMerkleTree(data [][]byte) *MerkleTree {
	var nodes []MerkleTreeNode

	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}

	for _, datum := range data {
		node := NewMerkleTreeNode(nil, nil, datum)
		nodes = append(nodes, *node)
	}

	for i := 0; i < len(data)/2; i++ {
		var newLevel []MerkleTreeNode

		for j := 0; j < len(nodes); j += 2 {
			node := NewMerkleTreeNode(&nodes[j], &nodes[j+1], nil)
			newLevel = append(newLevel, *node)
		}

		nodes = newLevel
	}

	mTree := MerkleTree{&nodes[0]}

	return &mTree
}