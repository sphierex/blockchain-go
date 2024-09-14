package main

import (
	"fmt"

	"github.com/sphierex/blockchain-go/internal"
)

func main() {
	bc := internal.NewBlockchain()
	bc.AddBlock("Send 1 BTC to Ivan")
	bc.AddBlock("Send 2 more BTC TO Ivan")

	fmt.Println(bc)
}
