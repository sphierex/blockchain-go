package internal

import (
	"strconv"
	"testing"
)

func TestBlockchain_AddBlock(t *testing.T) {

	bc, _ := NewBlockchain()
	err := bc.AddBlock("Send 1 BTC to Ivan")
	if err != nil {
		t.Fatal(err)
	}
	err = bc.AddBlock("Send 2 more BTC TO Ivan")
	if err != nil {
		t.Fatal(err)
	}

	bcIterator := bc.Iterator()

	for {
		block := bcIterator.Next()
		if len(block.PrevBlockHash) == 0 {
			break
		}

		t.Logf("Prev. hash: %x\n", block.PrevBlockHash)
		t.Logf("Data: %s\n", block.Data)
		t.Logf("Hash: %x\n", block.Hash)

		pow := NewProofOfWork(block)
		t.Logf("POW: %s\n", strconv.FormatBool(pow.Validate()))
	}
}
