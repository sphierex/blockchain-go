package blockchain

import (
	"encoding/hex"
	"errors"

	"go.etcd.io/bbolt"
)

const utxoBucket = "chain_state"

// UTXOSet represents UTXO set.
type UTXOSet struct {
	bc *Blockchain
}

// NewUTXOSet returns a UTXOSet.
func NewUTXOSet(bc *Blockchain) *UTXOSet {
	return &UTXOSet{
		bc: bc,
	}
}

// Rebuild rebuilds the UTXO set.
func (u *UTXOSet) Rebuild() error {
	db := u.bc.db
	bucketName := []byte(utxoBucket)

	err := db.Update(func(tx *bbolt.Tx) error {
		err := tx.DeleteBucket(bucketName)
		if err != nil && !errors.Is(err, bbolt.ErrBucketNotFound) {
			return err
		}

		_, err = tx.CreateBucket(bucketName)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	UTXO := u.bc.GetUTXO()
	err = db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket(bucketName)

		for id, v := range UTXO {
			key, _ := hex.DecodeString(id)
			err := b.Put(key, v.Serialize())
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// GetSpendableOutputs finds and returns unspent outputs to reference in inputs.
func (u *UTXOSet) GetSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	db := u.bc.db

	_ = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			txId := hex.EncodeToString(k)
			outs := DeserializeTxOutputs(v)
			for outIdx, out := range outs.Values {
				if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOutputs[txId] = append(unspentOutputs[txId], outIdx)
				}
			}
		}
		return nil
	})

	return accumulated, unspentOutputs
}

// GetUTXO finds UTXO for a public key hash.
func (u *UTXOSet) GetUTXO(pubKeyHash []byte) []TxOutput {
	var result []TxOutput
	db := u.bc.db
	_ = db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs := DeserializeTxOutputs(v)
			for _, out := range outs.Values {
				if out.IsLockedWithKey(pubKeyHash) {
					result = append(result, out)
				}
			}
		}

		return nil
	})

	return result
}

// TxCount returns the number of transactions in the UTXO set.
func (u *UTXOSet) TxCount() int {
	db := u.bc.db
	counter := 0

	err := db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			counter++
		}

		return nil
	})
	if err != nil {
		return 0
	}

	return counter
}

// Update updates the UTXO set with transactions from the Block
// is considered to be the tip of a blockchain
func (u *UTXOSet) Update(block *Block) error {
	db := u.bc.db

	err := db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		for _, tx := range block.Transactions {
			if tx.IsCoinbase() == false {
				for _, vin := range tx.Vin {
					updatedOuts := TxOutputs{}
					outsBytes := b.Get(vin.TxId)
					outs := DeserializeTxOutputs(outsBytes)

					for outIdx, out := range outs.Values {
						if outIdx != vin.Vout {
							updatedOuts.Values = append(updatedOuts.Values, out)
						}
					}

					if len(updatedOuts.Values) == 0 {
						err := b.Delete(vin.TxId)
						if err != nil {
							return err
						}
					} else {
						err := b.Put(vin.TxId, updatedOuts.Serialize())
						if err != nil {
							return err
						}
					}
				}
			}

			newOutputs := TxOutputs{}
			for _, out := range tx.Vout {
				newOutputs.Values = append(newOutputs.Values, out)
			}

			err := b.Put(tx.ID, newOutputs.Serialize())
			if err != nil {
				return err
			}
		}

		return nil
	})

	return err
}
