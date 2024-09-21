package merkle

import "crypto/sha256"

type MerkleTree struct {
	RootNode *MerkleNode
}

type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

func NewNode(left, right *MerkleNode, data []byte) *MerkleNode {
	node := MerkleNode{}

	if left == nil && right == nil {
		hash := sha256.Sum256(data)
		node.Data = hash[:]
	} else {
		prevHashes := append(left.Data, right.Data...)
		hash := sha256.Sum256(prevHashes)
		node.Data = hash[:]
	}

	node.Left = left
	node.Right = right

	return &node
}

func New(data [][]byte) *MerkleTree {
	var nodes []MerkleNode

	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}

	for _, v := range data {
		node := NewNode(nil, nil, v)
		nodes = append(nodes, *node)
	}

	for i := 0; i < len(data)/2; i++ {
		var newNodes []MerkleNode

		for j := 0; j < len(nodes); j += 2 {
			node := NewNode(&nodes[j], &nodes[j+1], nil)
			newNodes = append(newNodes, *node)
		}

		nodes = newNodes
	}

	mTree := MerkleTree{&nodes[0]}

	return &mTree
}
