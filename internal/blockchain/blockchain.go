package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"go.etcd.io/bbolt"
)

const (
	dbFilename          = "zblock/dbs/blockchain_%s.db"
	blocksBucket        = "blocks"
	genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
	latestHashKey       = "latest"
)

// Blockchain implements interactions with a DB.
type Blockchain struct {
	tip []byte
	db  *bbolt.DB
}

// CreateBlockchain creates a new blockchain DB.
func CreateBlockchain(node, address string) (*Blockchain, error) {
	var tip []byte
	dbPath := fmt.Sprintf(dbFilename, node)
	if _, err := os.Stat(dbPath); err != nil && os.IsExist(err) {
		return nil, fmt.Errorf("blockchain file %s exists", dbPath)
	}

	db, err := bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			return err
		}

		// create genesis block.
		cTx := NewCoinbaseTx(address, genesisCoinbaseData)
		genesis := NewGenesisBlock(cTx)
		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			return err
		}

		// set latest block hash.
		err = b.Put([]byte(latestHashKey), genesis.Hash)
		if err != nil {
			return err
		}

		tip = genesis.Hash
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &Blockchain{
		tip: tip,
		db:  db,
	}, nil
}

// NewBlockchain creates a new Blockchain with genesis Block.
func NewBlockchain(node string) (*Blockchain, error) {
	var tip []byte
	dbPath := fmt.Sprintf(dbFilename, node)
	if _, err := os.Stat(dbPath); err != nil && os.IsNotExist(err) {
		return nil, fmt.Errorf("blockchain file %s not exists", dbPath)
	}

	db, err := bbolt.Open(dbPath, 0600, nil)
	if err != nil {
		return nil, err
	}

	// get latest block hash.
	err = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte(latestHashKey))

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &Blockchain{
		tip: tip,
		db:  db,
	}, nil
}

func (bc *Blockchain) T() {
	fn := func(str string) ([]byte, error) {
		// 去掉方括号并按空格分割字符串
		str = strings.Trim(str, "[]")
		strArr := strings.Fields(str)

		// 创建一个字节数组
		byteArr := make([]byte, len(strArr))

		// 遍历分割后的字符串数组，并将每个值转换为字节
		for i, v := range strArr {
			num, err := strconv.Atoi(v) // 将字符串转换为整数
			if err != nil {
				return nil, fmt.Errorf("error converting string to int: %v", err)
			}
			byteArr[i] = byte(num) // 转换为 byte 类型
		}

		return byteArr, nil
	}
	hash, _ := fn("[0 0 136 118 158 229 113 93 75 61 103 21 34 81 170 93 213 109 48 18 69 195 228 44 134 132 157 199 203 157 143 188]")
	_ = bc.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(blocksBucket))

		return bucket.ForEach(func(k, v []byte) error {
			if bytes.Equal(k, hash) {
				log.Println(DeserializeBlock(v))
			}
			return nil
		})
	})
}

// Submit saves the block into the blockchain.
func (bc *Blockchain) Submit(block *Block) error {
	err := bc.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		exists := b.Get(block.Hash)
		if exists != nil {
			return nil
		}

		buf := block.Serialize()
		err := b.Put(block.Hash, buf)
		if err != nil {
			return err
		}

		latestHash := b.Get([]byte(latestHashKey))
		latestBlockData := b.Get(latestHash)
		latestBlock := DeserializeBlock(latestBlockData)
		if block.Height > latestBlock.Height {
			err = b.Put([]byte(latestHashKey), block.Hash)
			if err != nil {
				return err
			}
			bc.tip = block.Hash
		}

		return nil
	})

	return err
}

// Mine mines a new block with the provided transactions.
func (bc *Blockchain) Mine(txs []*Transaction) (*Block, error) {

	for _, tx := range txs {
		if bc.VerifyTx(tx) != true {
			return nil, fmt.Errorf("invlid transaction")
		}
	}

	latestBlock, err := bc.latest()
	if err != nil {
		return nil, err
	}

	block := NewBlock(txs, latestBlock.Hash, latestBlock.Height+1)

	err = bc.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(block.Hash, block.Serialize())
		if err != nil {
			return err
		}

		err = b.Put([]byte(latestHashKey), block.Hash)
		if err != nil {
			return err
		}

		bc.tip = block.Hash
		return nil
	})

	if err != nil {
		return nil, err
	}

	return block, nil
}

// latest get latest block.
func (bc *Blockchain) latest() (*Block, error) {
	return bc.getBlockByKey(bc.latestHash())
}

// latestHash get latest hash of the blockchain.
func (bc *Blockchain) latestHash() []byte {
	var latestHash []byte
	_ = bc.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		latestHash = b.Get([]byte(latestHashKey))

		return nil
	})

	return latestHash
}

// getBlockByKey retrieve blocks through cache key
func (bc *Blockchain) getBlockByKey(key []byte) (*Block, error) {
	var block *Block

	err := bc.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		data := b.Get(key)
		if data == nil {
			return errors.New("block is not found")
		}
		block = DeserializeBlock(data)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return block, nil
}

// GetBlockByHash finds a block by its hash and returns it
func (bc *Blockchain) GetBlockByHash(hash []byte) (Block, error) {
	block, err := bc.getBlockByKey(hash)

	return *block, err
}

// GetBestHeight returns the height of the latest block
func (bc *Blockchain) GetBestHeight() int {
	block, err := bc.latest()
	if err != nil {
		return 0
	}

	return block.Height
}

// GetBlockHashes returns a list of hashes of all the blocks in the chain
func (bc *Blockchain) GetBlockHashes() [][]byte {
	var blockHashes [][]byte

	err := bc.Foreach(func(block *Block) error {
		blockHashes = append(blockHashes, block.Hash)
		return nil
	})
	if err != nil {
		return nil
	}

	return blockHashes
}

// GetTransactionById get a transaction by its ID.
func (bc *Blockchain) GetTransactionById(id []byte) (Transaction, error) {
	var tx Transaction

	err := bc.Foreach(func(block *Block) error {
		for _, innerTx := range block.Transactions {
			if bytes.Equal(innerTx.ID, id) {
				tx = *innerTx
				break
			}
		}

		return nil
	})
	if err != nil {
		return tx, err
	}

	return tx, nil
}

// GetUTXO get all unspent transaction outputs and returns transactions with spent outputs removed.
func (bc *Blockchain) GetUTXO() map[string]TxOutputs {
	result := make(map[string]TxOutputs)
	spentTxOutputs := make(map[string][]int)

	_ = bc.Foreach(func(block *Block) error {
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)
		outputs:
			for outIdx, out := range tx.Vout {
				if spentTxOutputs[txID] != nil {
					for _, spentOutIdx := range spentTxOutputs[txID] {
						if spentOutIdx == outIdx {
							continue outputs
						}
					}
				}

				outs := result[txID]
				outs.Values = append(outs.Values, out)
				result[txID] = outs
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.TxId)
					spentTxOutputs[inTxID] = append(spentTxOutputs[inTxID], in.Vout)
				}
			}
		}

		return nil
	})

	return result
}

// SignTx signs inputs of a Transaction
func (bc *Blockchain) SignTx(tx *Transaction, privateKey ecdsa.PrivateKey) error {
	prevTxs := make(map[string]Transaction)

	for _, v := range tx.Vin {
		prevTx, err := bc.GetTransactionById(v.TxId)
		if err != nil {
			return err
		}

		prevTxs[hex.EncodeToString(prevTx.ID)] = prevTx
	}

	return tx.Sign(privateKey, prevTxs)
}

// VerifyTx verifies transaction input signatures
func (bc *Blockchain) VerifyTx(tx *Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}

	prevTxs := make(map[string]Transaction)
	for _, v := range tx.Vin {
		prevTx, err := bc.GetTransactionById(v.TxId)
		if err != nil {
			return false
		}
		prevTxs[hex.EncodeToString(prevTx.ID)] = prevTx
	}

	result, err := tx.Verify(prevTxs)

	return result && err == nil
}

func (bc *Blockchain) Foreach(fn func(*Block) error) error {
	i := &iterator{
		current: bc.tip,
		db:      bc.db,
	}

	for {
		block, err := i.Next()
		if err != nil {
			// none, break the loop.
			if errors.Is(err, ErrNoBlock) {
				break
			}
			return err
		}

		err = fn(block)
		if err != nil {
			return err
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return nil
}

var ErrNoBlock = errors.New("no more block")

// Iterator is used to iterate over blockchain blocks.
type iterator struct {
	current []byte
	db      *bbolt.DB
}

// Next returns block starting from the tip.
func (i *iterator) Next() (*Block, error) {
	var block *Block

	err := i.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		blockData := b.Get(i.current)
		if blockData == nil {
			return ErrNoBlock
		}
		block = DeserializeBlock(blockData)

		return nil
	})
	if err != nil {
		return nil, err
	}

	i.current = block.PrevBlockHash
	return block, nil
}
