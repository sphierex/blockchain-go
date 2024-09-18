package handlers

import (
	"fmt"
	"github.com/sphierex/blockchain-go/pkg"
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
	ba.rootCmd.AddCommand(ba.printCmd(), ba.balanceCmd(), ba.sendCmd(), ba.createBlockchainCmd(), ba.createAccount())

	return ba
}

func (ba *BlockchainApp) Execute() error {
	return ba.rootCmd.Execute()
}

func (ba *BlockchainApp) createBlockchainCmd() *cobra.Command {
	return &cobra.Command{
		Use: "create-blockchain",
		Run: func(cmd *cobra.Command, args []string) {
			address := args[0]
			internal.CreateBlockchain(address)
			fmt.Println("Done!")
		},
	}
}

func (ba *BlockchainApp) createAccount() *cobra.Command {
	return &cobra.Command{
		Use: "create-account",
		Run: func(cmd *cobra.Command, args []string) {
			wallets, _ := internal.NewWallets()
			address := wallets.Create()
			err := wallets.SaveToFile()
			if err != nil {
				fmt.Printf("wallet save error: %s\r\n", err)
				os.Exit(1)
			}

			fmt.Printf("Your new address: %s\n", address)
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
			pubKeyHash := pkg.Base58Decode([]byte(address))
			pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
			utxos := bc.FindUTXO(pubKeyHash)
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
