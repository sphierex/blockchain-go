package internal

import (
	"bytes"
	"encoding/gob"
	"github.com/sphierex/blockchain-go/pkg/merkletree"
	"time"
)

// Block represents a block in the blockchain.
type Block struct {
	Timestamp     int64          // 区块创建的时间
	Transactions  []*Transaction // 区块存储的信息
	PrevBlockHash []byte         // 前一个块的 Hash
	Hash          []byte         // 当前块的 Hash
	Nonce         int            // 区块的随机数
	Height        int
}

// NewBlock creates and returns Block.
func NewBlock(transactions []*Transaction, prevBlockHash []byte, height int) *Block {
	b := &Block{time.Now().Unix(), transactions, prevBlockHash, []byte{}, 0, height}

	pow := NewProofOfWork(b)
	nonce, hash := pow.Run()

	b.Hash = hash[:]
	b.Nonce = nonce

	return b
}

// NewGenesisBlock creates and returns genesis Block.
func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{}, 0)
}

func (b *Block) HashTransactions() []byte {
	var transactions [][]byte

	for _, tx := range b.Transactions {
		transactions = append(transactions, tx.Serialize())
	}

	mTree := merkletree.New(transactions)
	return mTree.RootNode.Data
}

func (b *Block) Serialize() []byte {
	var buf bytes.Buffer
	_ = gob.NewEncoder(&buf).Encode(b)

	return buf.Bytes()
}

func deserialize(buf []byte) *Block {
	var block Block
	_ = gob.NewDecoder(bytes.NewReader(buf)).Decode(&block)

	return &block
}
