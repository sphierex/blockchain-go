package internal

import (
	"crypto/sha256"
	"fmt"
	"math/big"
	"testing"
)

func Test_ProofOfWork_Lsh(t *testing.T) {
	data1 := []byte("I like donuts")
	data2 := []byte("I like donutsca07ca")

	target := big.NewInt(1)
	target.Lsh(target, uint(256-24))
	fmt.Printf("%x\n", sha256.Sum256(data1))
	fmt.Printf("%64x\n", target)
	fmt.Printf("%x\n", sha256.Sum256(data2))
}
