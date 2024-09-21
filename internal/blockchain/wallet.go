package blockchain

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"os"
)

const wallerFilename = "zblock/wallets/wallet_%s.dat"

// Wallet stores a collection of accounts.
type Wallet struct {
	Accounts map[string]*Account
}

// NewWallet creates Wallet and fills it from a file if it exists.
func NewWallet(node string) (*Wallet, error) {
	w := Wallet{}
	w.Accounts = make(map[string]*Account)
	err := w.Load(node)

	return &w, err
}

// NewAccount adds an Account to Wallet.
func (w *Wallet) NewAccount() string {
	account := NewAccount()
	w.Accounts[account.String()] = account

	return account.String()
}

// GetAddresses returns an array of addresses stored in the wallet file
func (w *Wallet) GetAddresses() []string {
	var addresses []string
	for address := range w.Accounts {
		addresses = append(addresses, address)
	}

	return addresses
}

// GetAccount returns an Account by its address
func (w *Wallet) GetAccount(address string) Account {
	return *w.Accounts[address]
}

// Load loads accounts from the file
func (w *Wallet) Load(node string) error {
	walletFile := fmt.Sprintf(wallerFilename, node)
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	content, err := ioutil.ReadFile(walletFile)
	if err != nil {
		return err
	}

	var wallet Wallet
	gob.Register(elliptic.P256())
	err = gob.NewDecoder(bytes.NewReader(content)).Decode(&wallet)
	if err != nil {
		return err
	}

	w.Accounts = wallet.Accounts

	return nil
}

// Save saves accounts to a file
func (w *Wallet) Save(node string) error {
	var buf bytes.Buffer
	walletFile := fmt.Sprintf(wallerFilename, node)

	gob.Register(elliptic.P256())
	err := gob.NewEncoder(&buf).Encode(w)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(walletFile, buf.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}
