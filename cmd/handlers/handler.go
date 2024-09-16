package handlers

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/sphierex/blockchain-go/internal"
)

type BlockchainApp struct {
	// bc      *internal.Blockchain
	rootCmd *cobra.Command
}

func NewBlockchainApp() *BlockchainApp {
	ba := &BlockchainApp{}

	ba.rootCmd = &cobra.Command{
		Use:   "blockchain",
		Short: "blockchain-go",
	}
	ba.rootCmd.AddCommand(ba.printCmd(), ba.balanceCmd(), ba.sendCmd(), ba.createCmd())

	return ba
}

func (ba *BlockchainApp) Execute() error {
	return ba.rootCmd.Execute()
}

func (ba *BlockchainApp) createCmd() *cobra.Command {
	return &cobra.Command{
		Use: "create",
		Run: func(cmd *cobra.Command, args []string) {
			address := args[0]
			internal.CreateBlockchain(address)
			fmt.Println("Done!")
		},
	}
}

func (ba *BlockchainApp) printCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "print",
		Short: "Print all block of chain",
		Run: func(cmd *cobra.Command, args []string) {
			bc, _ := internal.NewBlockchain()
			bci := bc.Iterator()

			for {
				block := bci.Next()
				fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
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

func (ba *BlockchainApp) balanceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "balance",
		Short: "Get balance of address",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			address := args[0]
			balance := 0

			bc, _ := internal.NewBlockchain()
			utxos := bc.FindUTXO(address)
			for _, out := range utxos {
				balance += out.Value
			}

			fmt.Printf("Balance of '%s': %d\n", address, balance)
		},
	}
}

func (ba *BlockchainApp) sendCmd() *cobra.Command {
	var from string
	var to string
	var amount int

	sendCmd := &cobra.Command{
		Use:   "send",
		Short: "Send amount of coins from FROM address to TO",
		Run: func(cmd *cobra.Command, args []string) {
			bc, _ := internal.NewBlockchain()
			tx, err := internal.NewUTXOTransaction(from, to, amount, bc)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			err = bc.MineBlock([]*internal.Transaction{tx})
			fmt.Println(err)
		},
	}

	sendCmd.Flags().StringVarP(&from, "from", "f", "", "Source wallet address")
	sendCmd.Flags().StringVarP(&to, "to", "t", "", "Destination wallet address")
	sendCmd.Flags().IntVarP(&amount, "amount", "a", 0, "Amount")

	_ = sendCmd.MarkFlagRequired("from")
	_ = sendCmd.MarkFlagRequired("to")
	_ = sendCmd.MarkFlagRequired("amount")

	return sendCmd
}
