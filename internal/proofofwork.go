package internal

import (
	"bytes"
	"crypto/sha256"
	"github.com/sphierex/blockchain-go/pkg"
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
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits)) // 1 << 212(256 - 44)

	pow := &ProofOfWork{b, target}

	return pow
}

// prepareData
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join([][]byte{
		pow.block.PrevBlockHash,
		pow.block.Data,
		pkg.IntToHex(pow.block.Timestamp),
		pkg.IntToHex(int64(targetBits)),
		pkg.IntToHex(int64(nonce)),
	}, []byte{})

	return data
}

// Run preforms a proof-of-work.
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	// fmt.Printf("Mining the block containg \"%s\"\n", pow.block.Data)

	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:])
		if hashInt.Cmp(pow.target) == -1 {
			// fmt.Printf("\r%x\n", hash)
			break
		} else {
			nonce++
		}
	}

	return nonce, hash[:]
}

// Validate validates block's POW.
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}
