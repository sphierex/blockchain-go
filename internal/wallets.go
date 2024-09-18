package internal

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"os"
)

const walletFile = "wallet.dat"

type Wallets struct {
	Accounts map[string]*Account
}

func NewWallets() (*Wallets, error) {
	wallets := Wallets{}
	wallets.Accounts = make(map[string]*Account)
	err := wallets.LoadFromFile()

	return &wallets, err
}

func (ws *Wallets) Create() string {
	account := NewAccount()
	address := fmt.Sprintf("%s", account.GetAddress())

	ws.Accounts[address] = account

	return address
}

func (ws *Wallets) GetAddresses() []string {
	var addresses []string

	for address := range ws.Accounts {
		addresses = append(addresses, address)
	}

	return addresses
}

func (ws *Wallets) GetAccount(address string) Account {
	return *ws.Accounts[address]
}

type EllipticCurveWrapper struct {
	Curve elliptic.Curve
}

func (ws *Wallets) LoadFromFile() error {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	content, err := os.ReadFile(walletFile)
	if err != nil {
		return err
	}

	var wallets Wallets
	gob.Register(elliptic.P256())
	err = gob.NewDecoder(bytes.NewReader(content)).Decode(&wallets)
	if err != nil {
		return err
	}

	ws.Accounts = wallets.Accounts

	return nil
}

func (ws *Wallets) SaveToFile() error {
	var content bytes.Buffer

	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		return err
	}

	err = os.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}
