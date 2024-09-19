package internal

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"os"
)

const walletFile = "wallet_%s.dat"

type Wallets struct {
	Accounts map[string]*Account
}

func NewWallets(nodeId string) (*Wallets, error) {
	wallets := Wallets{}
	wallets.Accounts = make(map[string]*Account)
	err := wallets.LoadFromFile(nodeId)

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

func (ws *Wallets) LoadFromFile(nodeId string) error {
	sWalletFile := fmt.Sprintf(walletFile, nodeId)
	if _, err := os.Stat(sWalletFile); os.IsNotExist(err) {
		return err
	}

	content, err := os.ReadFile(sWalletFile)
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

func (ws *Wallets) SaveToFile(nodeId string) error {
	sWalletFile := fmt.Sprintf(walletFile, nodeId)
	var content bytes.Buffer

	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		return err
	}

	err = os.WriteFile(sWalletFile, content.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}
