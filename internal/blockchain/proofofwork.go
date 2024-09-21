package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"math"
	"math/big"
)

var maxNonce = math.MaxInt64

const targetBits = 16

// ProofOfWork represents a proof-of-work.
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

// NewProofOfWork builds and returns a ProofOfWork.
func NewProofOfWork(block *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{
		block:  block,
		target: target,
	}

	return pow
}

// Run performs a proof-of-work
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	for nonce < maxNonce {
		data := pow.prepareData(nonce)

		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		}

		nonce++
	}

	return nonce, hash[:]
}

// prepareData Prepare data wait for performs.
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	ithFn := func(num int64) []byte {
		var buf bytes.Buffer
		_ = binary.Write(&buf, binary.BigEndian, num)
		return buf.Bytes()
	}

	data := bytes.Join([][]byte{
		pow.block.PrevBlockHash,
		pow.block.HashTransactions(),
		ithFn(pow.block.Timestamp),
		ithFn(int64(targetBits)),
		ithFn(int64(nonce)),
	}, []byte{})

	return data
}

// Validate validates block's PoW.
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	return hashInt.Cmp(pow.target) == -1
}
