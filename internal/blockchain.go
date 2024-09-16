package internal

import (
	"encoding/hex"
	"fmt"
	"os"

	bolt "go.etcd.io/bbolt"
)

const (
	dbPath              = "blochchain.db"
	blocksBucket        = "blocks"
	genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
)

// Blockchain keeps a sequence of Blocks.
type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

// NewBlockchain creates a new Blockchain with genesis Block.
func NewBlockchain() (*Blockchain, error) {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no existing blockchain found. Create one first")
	}

	var tip []byte
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("open invalid: %w", err)
	}

	// get latest block.
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("latest"))

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Blockchain{tip, db}, nil
}

func CreateBlockchain(address string) *Blockchain {
	if _, err := os.Stat(dbPath); os.IsExist(err) {
		os.Exit(1)
		return nil
	}

	var tip []byte
	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil
	}

	err = db.Update(func(tx *bolt.Tx) error {
		cBtx := NewCoinbaseTX(address, genesisCoinbaseData)
		genesis := NewGenesisBlock(cBtx)

		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			return err
		}

		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			return err
		}

		err = b.Put([]byte("latest"), genesis.Hash)
		if err != nil {
			return err
		}

		tip = genesis.Hash

		return nil
	})

	bc := Blockchain{tip, db}
	return &bc
}

// MineBlock mines a new block with the provided transactions.
func (bc *Blockchain) MineBlock(transactions []*Transaction) error {
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("latest"))

		return nil
	})
	if err != nil {
		return err
	}

	nb := NewBlock(transactions, lastHash)

	return bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(nb.Hash, nb.Serialize())
		if err != nil {
			return err
		}

		err = b.Put([]byte("latest"), nb.Hash)
		if err != nil {
			return err
		}

		bc.tip = nb.Hash

		return nil
	})
}

// FindUnspentTransactions returns a list of transaction containing unspent outputs.
func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
	var unspentTxs []Transaction
	spentTxOutputs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		b := bci.Next()
		for _, tx := range b.Transactions {
			txID := hex.EncodeToString(tx.ID)

		outputs:
			for idx, out := range tx.Vout {
				if spentTxOutputs[txID] != nil {
					for _, spentOut := range spentTxOutputs[txID] {
						if spentOut == idx {
							continue outputs
						}
					}
				}

				if out.CanBeUnlockedWith(address) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.TxId)
						spentTxOutputs[inTxID] = append(spentTxOutputs[inTxID], in.Vout)
					}
				}
			}

		}

		if len(b.Transactions) == 0 {
			break
		}
	}

	return unspentTxs
}

// FindUTXO finds and returns all unspent transaction outputs.
func (bc *Blockchain) FindUTXO(address string) []TxOutput {
	var utxos []TxOutput
	unspentTransactions := bc.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				utxos = append(utxos, out)
			}
		}
	}

	return utxos
}

func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTxs := bc.FindUnspentTransactions(address)
	accumulated := 0

work:

	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break work
				}
			}
		}
	}

	return accumulated, unspentOutputs
}

func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{
		currentHash: bc.tip,
		db:          bc.db,
	}

	return bci
}

type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

func (bi *BlockchainIterator) Next() *Block {
	var block *Block

	_ = bi.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		result := b.Get(bi.currentHash)
		// if result == nil {
		//	return fmt.Errorf("%s", "no more block")
		//}
		block = deserialize(result)
		return nil
	})

	bi.currentHash = block.PrevBlockHash

	return block
}
