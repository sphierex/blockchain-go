package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/sphierex/blockchain-go/pkg/base58"
)

const subsidy = 10

// TxInput represents a transaction input.
type TxInput struct {
	TxId      []byte
	Vout      int
	Signature []byte
	PubKey    []byte
}

// UsesKey checks whether the address initiated the transaction.
func (in *TxInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := HashPubKey(in.PubKey)

	return bytes.Equal(lockingHash, pubKeyHash)
}

// TxOutput represents a transaction output.
type TxOutput struct {
	Value      int
	PubKeyHash []byte
}

// Lock signs the output.
func (to *TxOutput) Lock(address []byte) {
	pubKeyHash := base58.Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	to.PubKeyHash = pubKeyHash
}

// IsLockedWithKey checks if the output can be used by the owner of the pubkey
func (to *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Equal(to.PubKeyHash, pubKeyHash)
}

// NewTxOutput create a new TXOutput.
func NewTxOutput(v int, address string) *TxOutput {
	txo := &TxOutput{Value: v, PubKeyHash: nil}
	txo.Lock([]byte(address))

	return txo
}

// TxOutputs collects TXOutput
type TxOutputs struct {
	Values []TxOutput
}

// Serialize serializes TXOutputs
func (to *TxOutputs) Serialize() []byte {
	var buf bytes.Buffer
	_ = gob.NewEncoder(&buf).Encode(to)

	return buf.Bytes()
}

// DeserializeTxOutputs deserializes TXOutputs
func DeserializeTxOutputs(data []byte) TxOutputs {
	var outputs TxOutputs
	_ = gob.NewDecoder(bytes.NewReader(data)).Decode(&outputs)

	return outputs
}

// Transaction represents a Bitcoin transaction.
type Transaction struct {
	ID   []byte
	Vin  []TxInput
	Vout []TxOutput
}

// NewCoinbaseTx creates a new coinbase transaction
func NewCoinbaseTx(to, data string) *Transaction {
	if data == "" {
		buf := make([]byte, 20)
		_, _ = rand.Read(buf)
		data = fmt.Sprintf("%x", buf)
	}

	txIn := TxInput{
		TxId:      []byte{},
		Vout:      -1,
		Signature: nil,
		PubKey:    []byte(data),
	}
	txOut := NewTxOutput(subsidy, to)
	tx := Transaction{
		ID:   nil,
		Vin:  []TxInput{txIn},
		Vout: []TxOutput{*txOut},
	}
	tx.ID = tx.Hash()

	return &tx
}

// IsCoinbase checks whether the transaction is coinbase.
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].TxId) == 0 && tx.Vin[0].Vout == -1
}

// Serialize returns a serialized Transaction.
func (tx *Transaction) Serialize() []byte {
	var buf bytes.Buffer
	_ = gob.NewEncoder(&buf).Encode(tx)

	return buf.Bytes()
}

// Hash returns the hash of the Transaction.
func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

// Sign signs each input of a Transaction.
func (tx *Transaction) Sign(privateKey ecdsa.PrivateKey, prevTxs map[string]Transaction) error {
	if tx.IsCoinbase() {
		return nil //
	}

	for _, v := range tx.Vin {
		if prevTxs[hex.EncodeToString(v.TxId)].ID == nil {
			return fmt.Errorf("%s", "previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	for id, v := range txCopy.Vin {
		prevTx := prevTxs[hex.EncodeToString(v.TxId)]
		txCopy.Vin[id].Signature = nil
		txCopy.Vin[id].PubKey = prevTx.Vout[v.Vout].PubKeyHash

		buf := fmt.Sprintf("%x\n", txCopy)
		r, s, err := ecdsa.Sign(rand.Reader, &privateKey, []byte(buf))
		if err != nil {
			return err
		}
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Vin[id].Signature = signature
		txCopy.Vin[id].PubKey = nil
	}

	return nil
}

// Verify verifies signatures of Transaction inputs.
func (tx *Transaction) Verify(prevTXs map[string]Transaction) (bool, error) {
	if tx.IsCoinbase() {
		return true, nil
	}

	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.TxId)].ID == nil {
			return false, fmt.Errorf("%s", "previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for id, v := range tx.Vin {
		prevTx := prevTXs[hex.EncodeToString(v.TxId)]
		txCopy.Vin[id].Signature = nil
		txCopy.Vin[id].PubKey = prevTx.Vout[v.Vout].PubKeyHash

		r := big.Int{}
		s := big.Int{}
		sigLen := len(v.Signature)
		r.SetBytes(v.Signature[:(sigLen / 2)])
		s.SetBytes(v.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(v.PubKey)
		x.SetBytes(v.PubKey[:(keyLen / 2)])
		y.SetBytes(v.PubKey[(keyLen / 2):])

		dataToVerify := fmt.Sprintf("%x\n", txCopy)

		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if ecdsa.Verify(&rawPubKey, []byte(dataToVerify), &r, &s) == false {
			return false, fmt.Errorf("%s", "tx vin verification failed")
		}
		txCopy.Vin[id].PubKey = nil
	}

	return true, nil
}

// TrimmedCopy creates a trimmed copy of Transaction to be used in signing.
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	for _, v := range tx.Vin {
		inputs = append(inputs, TxInput{
			TxId: v.TxId,
			Vout: v.Vout,
		})
	}

	for _, v := range tx.Vout {
		outputs = append(outputs, TxOutput{
			Value:      v.Value,
			PubKeyHash: v.PubKeyHash,
		})
	}

	txCopy := Transaction{ID: tx.ID, Vin: inputs, Vout: outputs}

	return txCopy
}

// String returns a human-readable representation of a transaction.
func (tx *Transaction) String() string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("   Transaction %x:\n", tx.ID))
	for i, input := range tx.Vin {
		builder.WriteString(fmt.Sprintf("     Input %d:\n", i))
		builder.WriteString(fmt.Sprintf("       TXID:      %x\n", input.TxId))
		builder.WriteString(fmt.Sprintf("       Out:       %d\n", input.Vout))
		builder.WriteString(fmt.Sprintf("       Signature: %x\n", input.Signature))
		builder.WriteString(fmt.Sprintf("       PubKey:    %x\n", input.PubKey))
	}

	for i, output := range tx.Vout {
		builder.WriteString(fmt.Sprintf("     Output %d:\n", i))
		builder.WriteString(fmt.Sprintf("       Value:  %d\n", output.Value))
		builder.WriteString(fmt.Sprintf("       Script: %x\n", output.PubKeyHash))
	}

	return builder.String()
}

// NewUTXOTransaction creates a new transaction.
func NewUTXOTransaction(account *Account, to string, amount int, UTXOSet *UTXOSet) (*Transaction, error) {
	var inputs []TxInput
	var outputs []TxOutput

	pubKeyHash := HashPubKey(account.PublicKey)
	quantity, validOutputs := UTXOSet.GetSpendableOutputs(pubKeyHash, amount)
	if quantity < amount {
		return nil, errors.New("not enough funds")
	}

	// Build a list of inputs.
	for id, outs := range validOutputs {
		txID, _ := hex.DecodeString(id)
		for _, out := range outs {
			input := TxInput{
				TxId:      txID,
				Vout:      out,
				Signature: nil,
				PubKey:    account.PublicKey,
			}
			inputs = append(inputs, input)
		}
	}

	from := account.String()
	outputs = append(outputs, *NewTxOutput(amount, to))
	if quantity > amount {
		outputs = append(outputs, *NewTxOutput(quantity-amount, from))
	}

	tx := &Transaction{
		ID:   nil,
		Vin:  inputs,
		Vout: outputs,
	}
	tx.ID = tx.Hash()

	err := UTXOSet.bc.SignTx(tx, account.PrivateKey)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// DeserializeTx deserializes a transaction.
func DeserializeTx(v []byte) Transaction {
	var tx Transaction
	_ = gob.NewDecoder(bytes.NewReader(v)).Decode(&tx)

	return tx
}
