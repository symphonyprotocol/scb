package block

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	// "encoding/gob"
	"github.com/symphonyprotocol/sutil/utils"
	"math"
	"log"
)

// type Content interface {
// 	CalculateHash() ([]byte, error)
// 	Equals(other Content) (bool, error)
// 	IsDup()(bool, error)
// 	SetDup(bool) Content
// }

type MerkleTree struct {
	Root       *Node
	merkleRoot []byte
	Leafs      []*Node
}

type Node struct {
	Parent *Node
	Left   *Node
	Right  *Node
	leaf   bool
	dup    bool
	// 是否是虚拟节点, 右子树构建时候除了第一个叶子节点外所有节点为虚拟
	virtual bool
	Hash   []byte
	C      BlockContent
}

type NodeShadow struct{
	Leaf bool
	Dup bool
	Virtual bool
	Hash [] byte
	C BlockContent
}

func (n *NodeShadow) Serialize() []byte {
	// var result bytes.Buffer
	// encoder := gob.NewEncoder(&result)

	// err := encoder.Encode(n)
	// if err != nil {
	// 	blockLogger.Error("Failed to serialize the node: %v", err)
	// 	panic(err)
	// }
	// resbytes  := result.Bytes()
	// return resbytes
	return utils.ObjToBytes(n)
}

func DeserializeNode(d []byte) *NodeShadow {
	var node NodeShadow = NodeShadow{}

	// decoder := gob.NewDecoder(bytes.NewReader(d))
	// err := decoder.Decode(&node)
	// if err != nil {
	// 	blockLogger.Error("Failed to deserialize the node: %v", err)
	// 	return nil
	// }
	utils.BytesToObj(d, &node)
	return &node
}

func (n *Node) verifyNode() ([]byte, error) {
	if n.leaf {
		return n.C.CalculateHash()
	}
	rightBytes, err := n.Right.verifyNode()
	if err != nil {
		return nil, err
	}

	leftBytes, err := n.Left.verifyNode()
	if err != nil {
		return nil, err
	}

	h := sha256.New()
	if _, err := h.Write(append(leftBytes, rightBytes...)); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

func (n *Node) calculateNodeHash() ([]byte, error) {
	if n.leaf {
		return n.C.CalculateHash()
	}

	h := sha256.New()
	if _, err := h.Write(append(n.Left.Hash, n.Right.Hash...)); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

func NewTree(cs []BlockContent) (*MerkleTree, error) {
	root, leafs, err := buildWithContent(cs)
	if err != nil {
		return nil, err
	}
	t := &MerkleTree{
		Root:       root,
		merkleRoot: root.Hash,
		Leafs:      leafs,
	}
	return t, nil
}

func buildWithContent(cs []BlockContent) (*Node, []*Node, error) {
	if len(cs) == 0 {
		return nil, nil, errors.New("error: cannot construct tree with no content")
	}
	var leafs []*Node
	for _, c := range cs {
		hash, err := c.CalculateHash()
		dup, err2 := c.IsDup()

		if err != nil {
			return nil, nil, err
		}
		if err2 != nil {
			return nil, nil, err2
		}

		leafs = append(leafs, &Node{
			Hash: hash,
			C:    c,
			leaf: true,
			virtual: dup,
		})
	}
	if len(leafs)%2 == 1 {
		duplicate := &Node{
			Hash: leafs[len(leafs)-1].Hash,
			C:    leafs[len(leafs)-1].C,
			leaf: true,
			dup:  true,
		}
		leafs = append(leafs, duplicate)
	}
	root, err := buildIntermediate(leafs)
	if err != nil {
		return nil, nil, err
	}

	return root, leafs, nil
}

func buildIntermediate(nl []*Node) (*Node, error) {
	var nodes []*Node
	for i := 0; i < len(nl); i += 2 {
		h := sha256.New()
		var left, right int = i, i + 1
		if i+1 == len(nl) {
			right = i
		}
		chash := append(nl[left].Hash, nl[right].Hash...)
		if _, err := h.Write(chash); err != nil {
			return nil, err
		}
		n := &Node{
			Left:  nl[left],
			Right: nl[right],
			Hash:  h.Sum(nil),
		}
		nodes = append(nodes, n)
		nl[left].Parent = n
		nl[right].Parent = n
		if len(nl) == 2 {
			return n, nil
		}
	}
	return buildIntermediate(nodes)
}

func (m *MerkleTree) MerkleRoot() []byte {
	return m.merkleRoot
}

func (m *MerkleTree) RebuildTree() error {
	var cs []BlockContent
	for _, c := range m.Leafs {
		cs = append(cs, c.C)
	}
	root, leafs, err := buildWithContent(cs)
	if err != nil {
		return err
	}
	m.Root = root
	m.Leafs = leafs
	m.merkleRoot = root.Hash
	return nil
}

func (m *MerkleTree) RebuildTreeWith(cs []BlockContent) error {
	root, leafs, err := buildWithContent(cs)
	if err != nil {
		return err
	}
	m.Root = root
	m.Leafs = leafs
	m.merkleRoot = root.Hash
	return nil
}

func (m *MerkleTree) VerifyTree() (bool, error) {
	calculatedMerkleRoot, err := m.Root.verifyNode()
	if err != nil {
		return false, err
	}

	if bytes.Compare(m.merkleRoot, calculatedMerkleRoot) == 0 {
		return true, nil
	}

	return false, nil
}

func (m *MerkleTree) VerifyContent(content BlockContent) (bool, error) {
	for _, l := range m.Leafs {
		ok, err := l.C.Equals(content)
		if err != nil {
			return false, err
		}

		if ok {
			currentParent := l.Parent
			for currentParent != nil {
				h := sha256.New()
				rightBytes, err := currentParent.Right.calculateNodeHash()
				if err != nil {
					return false, err
				}

				leftBytes, err := currentParent.Left.calculateNodeHash()
				if err != nil {
					return false, err
				}
				if currentParent.Left.leaf && currentParent.Right.leaf {
					if _, err := h.Write(append(leftBytes, rightBytes...)); err != nil {
						return false, err
					}
					if bytes.Compare(h.Sum(nil), currentParent.Hash) != 0 {
						return false, nil
					}
					currentParent = currentParent.Parent
				} else {
					if _, err := h.Write(append(leftBytes, rightBytes...)); err != nil {
						return false, err
					}
					if bytes.Compare(h.Sum(nil), currentParent.Hash) != 0 {
						return false, nil
					}
					currentParent = currentParent.Parent
				}
			}
			return true, nil
		}
	}
	return false, nil
}

func (m *MerkleTree) String() string {
	s := ""
	for _, l := range m.Leafs {
		s += fmt.Sprint(l)
		s += "\n"
	}
	return s
}

//获取节点的兄弟节点
func GetNodeBrother(node *Node) *Node{
	par_node := node.Parent
	if par_node == nil{
		return nil
	}

	if bytes.Compare(par_node.Left.Hash, node.Hash) == 0{
		return par_node.Right
	}else if bytes.Compare(par_node.Right.Hash, node.Hash) == 0{
		return par_node.Left
	}else{
		return nil
	}
}

//通过回溯兄弟节点获取Content的证明路径
func(m *MerkleTree) GetContentPath(content BlockContent) ([][]byte, error){
	var paths [][] byte

	var node *Node

	for _, l := range m.Leafs {
		ok, err := l.C.Equals(content)
		if err != nil {
			return nil , err
		}
		if ok{
			node = l
			break
		}
	}

	for true{
		if node.Parent == nil{
			break
		}
		brother := GetNodeBrother(node)
		if brother != nil{
			paths = append(paths, brother.Hash)
			node = node.Parent
		}else{
			break
		}
	}

	return paths, nil
}

//通过回溯兄弟节点获取节点的证明路径
func(m *MerkleTree) GetNodePath(node *Node) ([][]byte, error){
	var paths [][] byte

	for true{
		if node.Parent == nil{
			break
		}
		brother := GetNodeBrother(node)
		if brother != nil{
			paths = append(paths, brother.Hash)
			node = node.Parent
		}else{
			break
		}
	}

	return paths, nil
}

//merge 左右两颗结构一致的树为新树
func(left *MerkleTree) MergeTree(right *MerkleTree)(*MerkleTree, error){
	h := sha256.New()
	chash := append(left.Root.Hash, right.Root.Hash...)

	if _, err := h.Write(chash); err != nil {
		return nil, err
	}

	leafs := append(left.Leafs, right.Leafs...)
	
	n := &Node{
		Left: left.Root,
		Right: right.Root,
		Hash:  h.Sum(nil),
	}

	left.Root.Parent = n
	right.Root.Parent = n

	t := &MerkleTree{
		Root:       n,
		merkleRoot: n.Hash,
		Leafs:      leafs,
	}
	return t, nil
}

//树深度
func(m *MerkleTree) Depth() int64{
	var depth int64 = 0
	node := m.Root
	for{
		if node != nil{
			node = node.Left
			depth++
		}else{
			break
		}
	}
	return depth
}

//树叶子节点的数量
func(m *MerkleTree) LeafCount() int64{
	dep := m.Depth()
	res := math.Pow(2, float64(dep-1))
	return int64(res)
}

//寻找merkle的插入节点
func (m *MerkleTree) FindInsertPoint() *Node{
	for _, leaf := range m.Leafs {
		if leaf.dup{
			return leaf
		}else if leaf.virtual{
			return leaf
		}
	}
	return nil
}

/*merkle插入新的节点
	1. 若能找到插入节点, 更新此插入节点的content, 并更新回溯路径
	2. 若未能找到插入点，构建当前merkle 右子树 并与当前树合并
*/
func (m *MerkleTree) InsertContent(content BlockContent) *MerkleTree{
	position := m.FindInsertPoint()
	if position != nil{
		paths, _ := m.GetNodePath(position)
		m.UpdateNode(position, content, paths)
		m.merkleRoot = m.Root.Hash
		return m
	}else{
			leafCnt := m.LeafCount()
			var contents []BlockContent

			contents = append(contents, content)
			for i:=int64(0); i< leafCnt-1; i++{
				contentDup := content.SetDup(true)
				contents = append(contents, contentDup)
			}
			t2, _ := NewTree(contents)
			merged_tree, _ := m.MergeTree(t2)
			return merged_tree
	}
}

//根据证明路径更新节点
func(m *MerkleTree)UpdateNode(node *Node, content BlockContent, paths[][]byte){
	node.C = content
	node.dup = false
	node.virtual = false

	hash, _ := content.CalculateHash()
	node.Hash = hash

	var parent_node *Node = node.Parent
	var current_node *Node = node

	for len(paths) > 0{
		path := paths[0]
		paths = paths[1:]
		h := sha256.New()

		chash := append(path, current_node.Hash...)
		if _, err := h.Write(chash); err != nil {
			log.Panic(err)
		}
		parent_node.Hash = h.Sum(nil)
		current_node = parent_node
		parent_node = parent_node.Parent
	}
}

//广度优先序列化merkle树
func(m *MerkleTree)BreadthFirstSerialize() []byte {
	var result [][]byte
	var nodes []Node = []Node{*m.Root}

	for len(nodes) > 0 {
		node := nodes[0]
		nodes = nodes[1:]

		ns := & NodeShadow{
			Leaf: node.leaf,
			Dup: node.dup,
			Hash: node.Hash,
			Virtual: node.virtual,
			C: node.C,
		}

		result = append(result, ns.Serialize())
		if (node.Left != nil) {
			nodes = append(nodes, *node.Left)
		}
		if (node.Right != nil) {
			nodes = append(nodes, *node.Right)
		}
	}

	//把所有节点序列化

	return utils.ObjToBytes(result)

	// var result2 bytes.Buffer
	// encoder := gob.NewEncoder(&result2)

	// err := encoder.Encode(result)
	// if err == nil{
	// 	return result2.Bytes()
	// }

	// return nil
}

//反序列化数据为merkle树
func DeserializeNodeFromData(d []byte) *MerkleTree {

	//还原所有节点二进制数据
	var data [][]byte

	err := utils.BytesToObj(d, &data)
	if err != nil {
		blockLogger.Error("Failed to deserialize: %v", err)
		return nil
	}

	// decoder := gob.NewDecoder(bytes.NewReader(d))
	// err := decoder.Decode(&data)
	// if err != nil {
	// 	blockLogger.Error("Failed to deserialize: %v", err)
	// 	return nil
	// }	

	//根据节点还原树
	var tree *MerkleTree

    if len(data) == 0 {
        return nil
    }
    root := newNodeFromData(data[0])

    queue := make([]*Node, 1)
    queue[0] = root

	data = data[1:]

	var node *Node
	// var parent *Node
	var leafs [] *Node

    for len(data) > 0 && len(queue) > 0 {
		// parent = node
        node = queue[0]
        queue = queue[1:]

		// 父节点
		// node.Parent = parent

		// 左侧节点
		left := newNodeFromData(data[0])
		if left.leaf{
			leafs = append(leafs,left)
		}
		node.Left = left
		left.Parent = node
		
        if node.Left != nil {
            queue = append(queue, node.Left) 
        }
        data = data[1:]

        // 右侧节点
        if len(data) > 0 {
			right := newNodeFromData(data[0])
			if right.leaf{
				leafs = append(leafs,right)
			}
			node.Right = right
			right.Parent = node
            if node.Right != nil {
                queue = append(queue, node.Right)
            }
            data = data[1:]
        }
	}
	
	tree = &MerkleTree{
		Root : root,
		merkleRoot: root.Hash,
		Leafs: leafs,
	}

    return tree
}

func newNodeFromData(data [] byte) *Node {
		ns := DeserializeNode(data)
		node := &Node{
			leaf: ns.Leaf,
			dup: ns.Dup,
			Hash: ns.Hash,
			C: ns.C,
			virtual: ns.Virtual,
		}
		return node
}


func(m *MerkleTree) UpdateTree(changedAccounts []*Account, newAccounts []*Account) (*MerkleTree, error){
	if m == nil{
		var contents []BlockContent
		for _, account := range newAccounts {
			account_bytes := account.Serialize()
			contents = append(contents, BlockContent{
				X : account_bytes,
				Dup: false,
			})
		}
		tree, err := NewTree(contents)
		return tree, err
	}

	for _, account := range changedAccounts{
		idx := account.Index - 1
		updateNode := m.Leafs[idx]
		paths, _ := m.GetNodePath(updateNode)

		account_bytes := account.Serialize()
		m.UpdateNode(updateNode, BlockContent{
			X : account_bytes,
			Dup: false,
		}, paths)
	}
	m.merkleRoot = m.Root.Hash

	if len(newAccounts) > 0{
		var res_tree *MerkleTree
		for _, account := range newAccounts{
			account_bytes := account.Serialize()
	
			res_tree = m.InsertContent(
				BlockContent{
					X : account_bytes,
					Dup: false,
				})
		}
		return res_tree, nil
	}
	return m, nil
}

func (m *MerkleTree) DeserializeAccount() []*Account{
	var accounts [] *Account
	leafs := m.Leafs

	for _, leaf := range leafs{
		if leaf.dup || leaf.virtual{
			continue
		}
		account := DeserializeAccount(leaf.C.X)
		accounts = append(accounts, account)
	}
	return accounts
}
