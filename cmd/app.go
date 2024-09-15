package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/sphierex/blockchain-go/internal"
	"strconv"
)

type BlockchainApp struct {
	bc      *internal.Blockchain
	rootCmd *cobra.Command
}

func NewBlockchainApp() *BlockchainApp {
	ba := &BlockchainApp{}

	bc, _ := internal.NewBlockchain()
	ba.bc = bc

	ba.rootCmd = &cobra.Command{
		Use:   "blockchain",
		Short: "blockchain-go",
	}
	ba.rootCmd.AddCommand(ba.printCmd(), ba.generateCmd())

	return ba
}

func (ba *BlockchainApp) Execute() error {
	return ba.rootCmd.Execute()
}

func (ba *BlockchainApp) printCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "print",
		Short: "Print all block of chain",
		Run: func(cmd *cobra.Command, args []string) {
			bci := ba.bc.Iterator()
			for {
				block := bci.Next()
				fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
				fmt.Printf("Data: %s\n", block.Data)
				fmt.Printf("Hash: %x\n", block.Hash)
				pow := internal.NewProofOfWork(block)
				fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
				fmt.Println()
				if len(block.PrevBlockHash) == 0 {
					break
				}
			}
		},
	}
}

func (ba *BlockchainApp) generateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "generate",
		Short: "Generate a new block.",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			data := args[0]
			_ = ba.bc.AddBlock(data)
		},
	}
}
