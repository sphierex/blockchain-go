package handlers

import (
	"fmt"
	"github.com/sphierex/blockchain-go/pkg"
	"log"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/sphierex/blockchain-go/internal"
)

type BlockchainApp struct {
	// bc      *internal.Blockchain
	rootCmd *cobra.Command
	nodeId  string
}

func NewBlockchainApp() *BlockchainApp {
	ba := &BlockchainApp{}

	ba.rootCmd = &cobra.Command{
		Use:   "blockchain",
		Short: "blockchain-go",
	}
	ba.rootCmd.PersistentFlags().StringVarP(&ba.nodeId, "node", "", "3000", "global args")
	_ = ba.rootCmd.MarkFlagRequired("node")

	ba.rootCmd.AddCommand(ba.startNodeCmd(), ba.listAddresses(), ba.printCmd(), ba.balanceCmd(), ba.sendCmd(), ba.createBlockchainCmd(), ba.createAccount(), ba.rebuildIndex())

	return ba
}

func (ba *BlockchainApp) Execute() error {
	return ba.rootCmd.Execute()
}

// address nodeId
func (ba *BlockchainApp) createBlockchainCmd() *cobra.Command {
	return &cobra.Command{
		Use: "create-blockchain",
		Run: func(cmd *cobra.Command, args []string) {
			bc := internal.CreateBlockchain(args[0], ba.nodeId)
			fmt.Println("Done!")

			utxoSet := internal.UtxoSet{Blockchain: bc}
			utxoSet.ReIndex()
		},
	}
}

// nodeId
func (ba *BlockchainApp) createAccount() *cobra.Command {
	return &cobra.Command{
		Use: "create-account",
		Run: func(cmd *cobra.Command, args []string) {
			wallets, _ := internal.NewWallets(ba.nodeId)
			address := wallets.Create()
			err := wallets.SaveToFile(ba.nodeId)
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
			bc, _ := internal.NewBlockchain(ba.nodeId)
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
			balance := 0

			bc, _ := internal.NewBlockchain(ba.nodeId)
			utxoSet := internal.UtxoSet{Blockchain: bc}
			pubKeyHash := pkg.Base58Decode([]byte(args[0]))
			pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
			utxos := utxoSet.FindUtxo(pubKeyHash)
			for _, out := range utxos {
				balance += out.Value
			}

			fmt.Printf("Balance of '%s': %d\n", args[0], balance)
		},
	}
}

func (ba *BlockchainApp) sendCmd() *cobra.Command {
	var from string
	var to string
	var amount int
	var mineNow bool

	sendCmd := &cobra.Command{
		Use:   "send",
		Short: "Send amount of coins from FROM address to TO",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(111)
			fmt.Println(args)
			bc, err := internal.NewBlockchain(ba.nodeId)
			fmt.Println(bc, err)
			utxoSet := internal.UtxoSet{Blockchain: bc}

			fmt.Println(bc, utxoSet)

			wallets, err := internal.NewWallets(ba.nodeId)
			if err != nil {
				log.Panic(err)
			}
			wallet := wallets.GetAccount(from)

			tx, err := internal.NewUTXOTransaction(&wallet, to, amount, &utxoSet)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if mineNow {
				cbTx := internal.NewCoinbaseTX(from, "")
				txs := []*internal.Transaction{cbTx, tx}
				newBlock, _ := bc.MineBlock(txs)
				utxoSet.Update(newBlock)
			} else {
				internal.SendTx(internal.GetKnownNodes()[0], tx)
			}

			fmt.Println(err)
		},
	}

	sendCmd.Flags().StringVarP(&from, "from", "f", "", "Source wallet address")
	sendCmd.Flags().StringVarP(&to, "to", "t", "", "Destination wallet address")
	sendCmd.Flags().IntVarP(&amount, "amount", "a", 0, "Amount")
	sendCmd.Flags().BoolVarP(&mineNow, "mine", "m", false, "Mint now")

	_ = sendCmd.MarkFlagRequired("from")
	_ = sendCmd.MarkFlagRequired("to")
	_ = sendCmd.MarkFlagRequired("amount")
	_ = sendCmd.MarkFlagRequired("mine")

	return sendCmd
}

func (ba *BlockchainApp) rebuildIndex() *cobra.Command {
	return &cobra.Command{
		Use: "rebuild-index",
		Run: func(cmd *cobra.Command, args []string) {
			bc, _ := internal.NewBlockchain(ba.nodeId)
			utxoSet := internal.UtxoSet{Blockchain: bc}
			utxoSet.ReIndex()

			count := utxoSet.CountTransactions()
			fmt.Printf("There are %d transactions in the utxo set. \n", count)
		},
	}
}

func (ba *BlockchainApp) listAddresses() *cobra.Command {
	return &cobra.Command{
		Use: "list-addresses",
		Run: func(cmd *cobra.Command, args []string) {
			wallets, _ := internal.NewWallets(ba.nodeId)
			addresses := wallets.GetAddresses()
			for _, address := range addresses {
				fmt.Println(address)
			}
		},
	}
}

func (ba *BlockchainApp) startNodeCmd() *cobra.Command {
	return &cobra.Command{
		Use: "start-node",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args[0]) > 0 {
				if internal.ValidateAddress(args[0]) {
					fmt.Println("Mining is on. Address to receive rewards: ", args[0])
				}
			}
			internal.StartServer(ba.nodeId, args[0])
		},
	}
}
