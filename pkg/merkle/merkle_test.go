package merkle

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_MerkleNode(t *testing.T) {
	data := [][]byte{
		[]byte("node1"),
		[]byte("node2"),
		[]byte("node3"),
	}

	n1 := NewNode(nil, nil, data[0])
	n2 := NewNode(nil, nil, data[1])
	n3 := NewNode(nil, nil, data[2])
	n4 := NewNode(nil, nil, data[2])

	// Level 2
	n5 := NewNode(n1, n2, nil)
	n6 := NewNode(n3, n4, nil)

	// Level 3
	n7 := NewNode(n5, n6, nil)

	assert.Equal(
		t,
		"64b04b718d8b7c5b6fd17f7ec221945c034cfce3be4118da33244966150c4bd4",
		hex.EncodeToString(n5.Data),
		"Level 1 hash 1 is correct",
	)
	assert.Equal(
		t,
		"08bd0d1426f87a78bfc2f0b13eccdf6f5b58dac6b37a7b9441c1a2fab415d76c",
		hex.EncodeToString(n6.Data),
		"Level 1 hash 2 is correct",
	)
	assert.Equal(
		t,
		"4e3e44e55926330ab6c31892f980f8bfd1a6e910ff1ebc3f778211377f35227e",
		hex.EncodeToString(n7.Data),
		"Root hash is correct",
	)
}

func Test_MerkleTree(t *testing.T) {
	data := [][]byte{
		[]byte("node1"),
		[]byte("node2"),
		[]byte("node3"),
	}
	// Level 1
	n1 := NewNode(nil, nil, data[0])
	n2 := NewNode(nil, nil, data[1])
	n3 := NewNode(nil, nil, data[2])
	n4 := NewNode(nil, nil, data[2])

	// Level 2
	n5 := NewNode(n1, n2, nil)
	n6 := NewNode(n3, n4, nil)

	// Level 3
	n7 := NewNode(n5, n6, nil)

	rootHash := fmt.Sprintf("%x", n7.Data)
	mTree := New(data)

	assert.Equal(t, rootHash, fmt.Sprintf("%x", mTree.RootNode.Data), "Merkle tree root hash is correct")
}
