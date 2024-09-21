package app

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/sphierex/blockchain-go/internal/blockchain"
	"github.com/sphierex/blockchain-go/pkg/base58"
	"log"
	"os"
	"strconv"
)

type App struct {
	rootCmd *cobra.Command

	node string
}

func New() *App {
	app := &App{}
	app.init()

	return app
}

func (a *App) init() {
	_ = os.MkdirAll("zblock/dbs", 0644)
	_ = os.MkdirAll("zblock/wallets", 0644)

	rootCmd := &cobra.Command{
		Use: "go-blockchain",
	}

	rootCmd.PersistentFlags().StringVarP(&a.node, "node", "n", os.Getenv("NODE"), "")
	_ = rootCmd.MarkFlagRequired("node")

	rootCmd.AddCommand(
		a.createChainCmd(),
		a.createWalletCmd(),
		a.printChainCmd(),
		a.printAddressCmd(),
		a.getBalanceCmd(),
		a.rebuildChainStateCmd(),
		a.transformCmd(),
		a.startServerCmd(),
	)
	a.rootCmd = rootCmd
}

func (a *App) createChainCmd() *cobra.Command {
	var address string

	createChainCmd := &cobra.Command{
		Use:   "create-chain",
		Short: "Create a blockchain and send genesis block reward to address",
		Run: func(cmd *cobra.Command, args []string) {
			if !blockchain.ValidateAddress(address) {
				log.Println("address is not valid")
				os.Exit(1)
			}

			bc, err := blockchain.CreateBlockchain(a.node, address)
			if err != nil {
				log.Println(err)
				os.Exit(1)
			}

			err = blockchain.NewUTXOSet(bc).Rebuild()
			if err != nil {
				cmd.Println(err)
				os.Exit(1)
			}

			cmd.Println("done")
		},
	}

	createChainCmd.Flags().StringVarP(&address, "address", "", "", "The address to send genesis block reward to")
	_ = createChainCmd.MarkFlagRequired("address")

	return createChainCmd
}

func (a *App) createWalletCmd() *cobra.Command {
	return &cobra.Command{
		Use: "create-wallet",
		Run: func(cmd *cobra.Command, args []string) {
			wallet, _ := blockchain.NewWallet(a.node)
			account := wallet.NewAccount()

			err := wallet.Save(a.node)
			if err != nil {
				cmd.Printf("save account: %s", err)
				os.Exit(1)
			}

			cmd.Printf("New address: %s\n", account)
		},
	}
}

func (a *App) printChainCmd() *cobra.Command {
	return &cobra.Command{
		Use: "print-chain",
		Run: func(cmd *cobra.Command, args []string) {
			bc, err := blockchain.NewBlockchain(a.node)
			if err != nil {
				cmd.Println(err)
				os.Exit(1)
			}

			_ = bc.Foreach(func(block *blockchain.Block) error {
				fmt.Printf("============ Block %x ============\n", block.Hash)
				fmt.Printf("Height: %d\n", block.Height)
				fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)
				pow := blockchain.NewProofOfWork(block)
				fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
				for _, tx := range block.Transactions {
					fmt.Println(tx)
				}
				fmt.Printf("\n")

				return nil
			})
		},
	}
}

func (a *App) printAddressCmd() *cobra.Command {
	return &cobra.Command{
		Use: "print-addresses",
		Run: func(cmd *cobra.Command, args []string) {
			wallet, err := blockchain.NewWallet(a.node)
			if err != nil {
				cmd.Println(err)
				os.Exit(1)
			}
			addresses := wallet.GetAddresses()
			for i, address := range addresses {
				fmt.Printf("%d: %s\r\n", i+1, address)
			}
			fmt.Printf("total address: %d\r\n", len(addresses))
		},
	}
}

func (a *App) getBalanceCmd() *cobra.Command {
	var address string

	getBalanceCmd := &cobra.Command{
		Use:   "get-balance",
		Short: "Create a blockchain and send genesis block reward to address",
		Run: func(cmd *cobra.Command, args []string) {
			if !blockchain.ValidateAddress(address) {
				cmd.Println("address is not valid")
				os.Exit(1)
			}

			bc, err := blockchain.NewBlockchain(a.node)
			if err != nil {
				cmd.Println(err)
				os.Exit(1)
			}

			UTXOSet := blockchain.NewUTXOSet(bc)
			balance := 0

			pubKeyHash := base58.Decode([]byte(address))
			pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]

			UTXOs := UTXOSet.GetUTXO(pubKeyHash)
			for _, out := range UTXOs {
				balance += out.Value
			}

			fmt.Printf("Balance of '%s': %d\n", address, balance)
		},
	}

	getBalanceCmd.Flags().StringVarP(&address, "address", "", "", "The address to send genesis block reward to")
	_ = getBalanceCmd.MarkFlagRequired("address")

	return getBalanceCmd
}

func (a *App) rebuildChainStateCmd() *cobra.Command {
	return &cobra.Command{
		Use: "rebuild-chain-state",
		Run: func(cmd *cobra.Command, args []string) {
			bc, err := blockchain.NewBlockchain(a.node)
			if err != nil {
				cmd.Println(err)
				os.Exit(1)
			}

			UTXOSet := blockchain.NewUTXOSet(bc)
			if err := UTXOSet.Rebuild(); err != nil {
				cmd.Println(err)
				os.Exit(1)
			}

			count := UTXOSet.TxCount()
			cmd.Printf("There are %d transaction in the UTXO set.\n", count)
		},
	}
}

func (a *App) transformCmd() *cobra.Command {
	var from, to string
	var amount int
	var mine bool

	transformCmd := &cobra.Command{
		Use: "transfer",
		Run: func(cmd *cobra.Command, args []string) {

			if !blockchain.ValidateAddress(from) {
				cmd.Println("sender address is not valid")
				os.Exit(1)
			}

			if !blockchain.ValidateAddress(to) {
				cmd.Println("recipient address is not valid")
				os.Exit(1)
			}

			bc, err := blockchain.NewBlockchain(a.node)
			if err != nil {
				cmd.Println(err)
				os.Exit(1)
			}

			UTXOSet := blockchain.NewUTXOSet(bc)

			wallet, err := blockchain.NewWallet(a.node)
			if err != nil {
				cmd.Println(err)
				os.Exit(1)
			}

			account := wallet.GetAccount(from)

			tx, err := blockchain.NewUTXOTransaction(&account, to, amount, UTXOSet)
			if err != nil {
				cmd.Println(err)
				os.Exit(1)
			}

			if mine {
				cTx := blockchain.NewCoinbaseTx(from, "")
				txs := []*blockchain.Transaction{cTx, tx}

				nBlock, err := bc.Mine(txs)
				if err != nil {
					cmd.Println(err)
					os.Exit(1)
				}

				err = UTXOSet.Update(nBlock)
				if err != nil {
					cmd.Println(err)
					os.Exit(1)
				}
			} else {
				s := blockchain.NewServerWithBlockchain(bc, a.node, "")
				s.SendTx(tx)
			}

			cmd.Println("success")
		},
	}
	transformCmd.Flags().StringVarP(&from, "from", "", "", "")
	transformCmd.Flags().StringVarP(&to, "to", "", "", "")
	transformCmd.Flags().IntVarP(&amount, "amount", "", 0, "")
	transformCmd.Flags().BoolVarP(&mine, "mine", "", false, "")
	_ = transformCmd.MarkFlagRequired("from")
	_ = transformCmd.MarkFlagRequired("to")
	_ = transformCmd.MarkFlagRequired("amount")

	return transformCmd
}

func (a *App) startServerCmd() *cobra.Command {
	var address string

	getBalanceCmd := &cobra.Command{
		Use: "start-server",
		Run: func(cmd *cobra.Command, args []string) {
			if !blockchain.ValidateAddress(address) {
				cmd.Println("address is not valid")
				os.Exit(1)
			}

			s := blockchain.NewServer(a.node, address)
			if err := s.Start(); err != nil {
				cmd.Println(err)
				os.Exit(1)
			}
		},
	}

	getBalanceCmd.Flags().StringVarP(&address, "address", "", "", "The address to send genesis block reward to")
	_ = getBalanceCmd.MarkFlagRequired("address")

	return getBalanceCmd
}

func (a *App) Execute() error {
	return a.rootCmd.Execute()
}
