package internal

import (
	"encoding/hex"
	bolt "go.etcd.io/bbolt"
)

const utxoBucket = "utxo:set"

type UtxoSet struct {
	Blockchain *Blockchain
}

func (u UtxoSet) ReIndex() {
	db := u.Blockchain.db
	bucketName := []byte(utxoBucket)

	_ = db.Update(func(tx *bolt.Tx) error {
		_ = tx.DeleteBucket(bucketName)
		_, _ = tx.CreateBucket(bucketName)

		return nil
	})

	utxo := u.Blockchain.FindUTXO()

	_ = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		for txID, outs := range utxo {
			key, _ := hex.DecodeString(txID)
			_ = b.Put(key, outs.Serialize())
		}

		return nil
	})
}

func (u UtxoSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	db := u.Blockchain.db

	_ = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			outs := DeserializeOutputs(v)

			for i, vv := range outs.Values {
				if vv.IsLockedWithKey(pubKeyHash) && accumulated < amount {
					accumulated += vv.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], i)
				}
			}
		}

		return nil
	})

	return accumulated, unspentOutputs
}

func (u UtxoSet) FindUtxo(pubKeyHash []byte) []TxOutput {
	var utxos []TxOutput
	db := u.Blockchain.db

	_ = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs := DeserializeOutputs(v)

			for _, vv := range outs.Values {
				if vv.IsLockedWithKey(pubKeyHash) {
					utxos = append(utxos, vv)
				}
			}
		}

		return nil
	})

	return utxos
}

func (u UtxoSet) CountTransactions() int {
	db := u.Blockchain.db
	counter := 0

	_ = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			counter++
		}

		return nil
	})

	return counter
}

func (u UtxoSet) Update(block *Block) {
	db := u.Blockchain.db

	_ = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))

		for _, tx := range block.Transactions {

			if !tx.IsCoinbase() {
				for _, vin := range tx.Vin {
					updateOuts := TxOutputs{}
					outsBytes := b.Get(vin.TxID)
					outs := DeserializeOutputs(outsBytes)

					for outIndex, out := range outs.Values {
						if outIndex != vin.Vout {
							updateOuts.Values = append(updateOuts.Values, out)
						}
					}

					if len(updateOuts.Values) == 0 {
						_ = b.Delete(vin.TxID)
					} else {
						_ = b.Put(vin.TxID, updateOuts.Serialize())
					}
				}
			}

			newOutputs := TxOutputs{}
			for _, out := range tx.Vout {
				newOutputs.Values = append(newOutputs.Values, out)
			}

			_ = b.Put(tx.ID, newOutputs.Serialize())
		}

		return nil
	})
}
