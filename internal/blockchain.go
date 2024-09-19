package internal

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"

	bolt "go.etcd.io/bbolt"
)

const (
	dbPath              = "blochchain_%s.db"
	blocksBucket        = "blocks"
	genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
)

// Blockchain keeps a sequence of Blocks.
type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

// NewBlockchain creates a new Blockchain with genesis Block.
func NewBlockchain(nodeID string) (*Blockchain, error) {
	subDbPath := fmt.Sprintf(dbPath, nodeID)
	if _, err := os.Stat(subDbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no existing blockchain found. Create one first")
	}

	var tip []byte
	db, err := bolt.Open(subDbPath, 0600, nil)
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

func CreateBlockchain(address string, nodeId string) *Blockchain {
	subDbPath := fmt.Sprintf(dbPath, nodeId)
	if _, err := os.Stat(subDbPath); os.IsExist(err) {
		os.Exit(1)
		return nil
	}

	var tip []byte
	db, err := bolt.Open(subDbPath, 0600, nil)
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
func (bc *Blockchain) MineBlock(transactions []*Transaction) (*Block, error) {
	var lastHash []byte
	var lastHeight int

	for _, tx := range transactions {
		if bc.VerifyTransaction(tx) != true {
			log.Panic("ERROR: Invalid transaction")
		}
	}

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("latest"))

		blockData := b.Get(lastHash)
		block := deserialize(blockData)

		lastHeight = block.Height

		return nil
	})
	if err != nil {
		return nil, err
	}

	nb := NewBlock(transactions, lastHash, lastHeight+1)

	return nb, bc.db.Update(func(tx *bolt.Tx) error {
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

func (bc *Blockchain) AddBlock(block *Block) {
	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		blockInDb := b.Get(block.Hash)

		if blockInDb != nil {
			return nil
		}

		blockData := block.Serialize()
		err := b.Put(block.Hash, blockData)
		if err != nil {
			log.Panic(err)
		}

		lastHash := b.Get([]byte("l"))
		lastBlockData := b.Get(lastHash)
		lastBlock := deserialize(lastBlockData)

		if block.Height > lastBlock.Height {
			err = b.Put([]byte("l"), block.Hash)
			if err != nil {
				log.Panic(err)
			}
			bc.tip = block.Hash
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

func (bc *Blockchain) GetBlock(blockHash []byte) (Block, error) {
	var block Block

	err := bc.db.View(func(t *bolt.Tx) error {
		b := t.Bucket([]byte(blocksBucket))
		blockData := b.Get(blockHash)

		if blockData == nil {
			return errors.New("block is no found")
		}

		block = *deserialize(blockData)

		return nil
	})

	if err != nil {
		return block, err
	}

	return block, nil
}

func (bc *Blockchain) GetBlockHashes() [][]byte {
	var blocks [][]byte
	bci := bc.Iterator()

	for {
		block := bci.Next()

		blocks = append(blocks, block.Hash)

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return blocks
}

func (bc *Blockchain) GetBestHeight() int {
	var lastBlock Block

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash := b.Get([]byte("l"))
		blockData := b.Get(lastHash)
		lastBlock = *deserialize(blockData)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return lastBlock.Height
}

// FindUnspentTransactions returns a list of transaction containing unspent outputs.
func (bc *Blockchain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
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

				if out.IsLockedWithKey(pubKeyHash) {
					unspentTxs = append(unspentTxs, *tx)
				}
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.UsesKey(pubKeyHash) {
						inTxID := hex.EncodeToString(in.TxID)
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
func (bc *Blockchain) FindUTXO() map[string]TxOutputs {
	utxos := make(map[string]TxOutputs)
	spentTxs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)
		outputs:
			for outIndex, out := range tx.Vout {
				if spentTxs[txID] != nil {
					for _, spentOutIdx := range spentTxs[txID] {
						if spentOutIdx == outIndex {
							continue outputs
						}
					}
				}

				outs := utxos[txID]
				outs.Values = append(outs.Values, out)
				utxos[txID] = outs
			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.TxID)
					spentTxs[inTxID] = append(spentTxs[inTxID], in.Vout)
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return utxos
}

func (bc *Blockchain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTxs := bc.FindUnspentTransactions(pubKeyHash)
	accumulated := 0

work:

	for _, tx := range unspentTxs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
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

func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()
		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("transaction is not found")
}

func (bc *Blockchain) SignTransaction(tx *Transaction, privateKey ecdsa.PrivateKey) {
	prevTxs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTx, _ := bc.FindTransaction(vin.TxID)
		prevTxs[hex.EncodeToString(prevTx.ID)] = prevTx
	}

	tx.Sign(privateKey, prevTxs)
	// return tx.Verify(prevTxs)
}

func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	prevTxs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTx, _ := bc.FindTransaction(vin.TxID)
		prevTxs[hex.EncodeToString(prevTx.ID)] = prevTx
	}

	return tx.Verify(prevTxs)
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

// DeserializeTransaction deserializes a transaction
func DeserializeTransaction(data []byte) Transaction {
	var transaction Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&transaction)
	if err != nil {
		log.Panic(err)
	}

	return transaction
}
