package internal

import (
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
}

// NewBlock creates and returns Block.
func NewBlock(data string, prevBlockHash []byte) *Block {
	b := &Block{
		Timestamp:     time.Now().Unix(),
		Data:          []byte(data),
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
	}
	b.SetHash()

	return b
}

// SetHash calculates and sets the block hash.
func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))

	var headers []byte
	headers = append(headers, b.PrevBlockHash...)
	headers = append(headers, b.Data...)
	headers = append(headers, timestamp...)

	hash := sha256.Sum256(headers)
	b.Hash = hash[:]
}

// NewGenesisBlock creates and returns genesis Block.
func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}
