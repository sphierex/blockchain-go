package internal

import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"time"
)

// Block represents a block in the blockchain.
type Block struct {
	Timestamp     int64  // 区块创建的时间
	Data          []byte // 区块存储的信息
	PrevBlockHash []byte // 前一个块的 Hash
	Hash          []byte // 当前块的 Hash
	Nonce         int    // 区块的随机数
}

// NewBlock creates and returns Block.
func NewBlock(data string, prevBlockHash []byte) *Block {
	b := &Block{
		Timestamp:     time.Now().Unix(),
		Data:          []byte(data),
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
		Nonce:         0,
	}

	pow := NewProofOfWork(b)
	nonce, hash := pow.Run()

	b.Hash = hash[:]
	b.Nonce = nonce

	return b
}

// SetHash calculates and sets the block hash.
func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))

	headers := bytes.Join([][]byte{
		b.PrevBlockHash,
		b.Data,
		timestamp,
	}, []byte{})

	hash := sha256.Sum256(headers)
	b.Hash = hash[:]
}

// NewGenesisBlock creates and returns genesis Block.
func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}
