package blockchain

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/sphierex/blockchain-go/pkg/merkle"
)

// Block represents a block in the blockchain.
type Block struct {
	Timestamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
	Height        int
}

// NewBlock creates and returns Block.
func NewBlock(transactions []*Transaction, prevBlockHash []byte, height int) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Transactions:  transactions,
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
		Nonce:         0,
		Height:        height,
	}

	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

// NewGenesisBlock creates and returns genesis Block.
func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{}, 0)
}

// HashTransactions returns a hash of the transactions in the block.
func (block *Block) HashTransactions() []byte {
	var transactions [][]byte

	for _, tx := range block.Transactions {
		transactions = append(transactions, tx.Serialize())
	}
	mTree := merkle.New(transactions)

	return mTree.RootNode.Data
}

// Serialize serializes the block.
func (block *Block) Serialize() []byte {
	var buf bytes.Buffer
	_ = gob.NewEncoder(&buf).Encode(block)

	return buf.Bytes()
}

// DeserializeBlock deserializes a block.
func DeserializeBlock(v []byte) *Block {
	var block Block
	_ = gob.NewDecoder(bytes.NewReader(v)).Decode(&block)

	return &block
}
