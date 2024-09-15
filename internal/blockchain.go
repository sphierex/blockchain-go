package internal

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

const (
	dbPath       = "blochchain.db"
	blocksBucket = "blocks"
)

// Blockchain keeps a sequence of Blocks.
type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

// NewBlockchain creates a new Blockchain with genesis Block.
func NewBlockchain() (*Blockchain, error) {
	var tip []byte

	db, err := bolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("open invalid: %w", err)
	}

	// get latest block.
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if b == nil {
			genesis := NewGenesisBlock()

			b, err = tx.CreateBucket([]byte(blocksBucket))
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
		}

		tip = b.Get([]byte("latest"))

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &Blockchain{tip, db}, nil
}

// AddBlock saves provided data as a block in the blockchain.
func (bc *Blockchain) AddBlock(data string) error {
	var lastHash []byte

	if err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("latest"))

		return nil
	}); err != nil {
		return err
	}

	newBlock := NewBlock(data, lastHash)

	if err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			return err
		}
		err = b.Put([]byte("latest"), newBlock.Hash)
		if err != nil {
			return err
		}
		bc.tip = newBlock.Hash

		return nil
	}); err != nil {
		return err
	}

	return nil
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
		if result == nil {
			return fmt.Errorf("%s", "no more block")
		}

		block = deserialize(result)
		return nil
	})

	//if err != nil {
	//	return nil
	//}

	bi.currentHash = block.PrevBlockHash

	return block
}
