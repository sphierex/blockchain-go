package internal

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"log"
	"math/big"
	"testing"
)

func TestWallet_GetAddress(t *testing.T) {
	account := NewAccount()
	t.Logf("%s", account.GetAddress())

	jsonData, err := json.Marshal(account)
	if err != nil {
		t.Fatalf("json marshal err: %s", err)
	}
	t.Logf("new account marshal data: %s\r\n", jsonData)

	var unmarshalAccount Account
	err = json.Unmarshal(jsonData, &unmarshalAccount)
	if err != nil {
		t.Fatalf("json unmarshal err: %s", err)
	}
	t.Logf("unmarshal account: %v.\r\n", unmarshalAccount)
}

type P256PublicKey struct {
	X, Y *big.Int
}

func (p *P256PublicKey) Key() *ecdsa.PublicKey {
	return &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     p.X,
		Y:     p.Y,
	}
}

func TestAccount_Curve(t *testing.T) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	pub := &privKey.PublicKey
	key := P256PublicKey{
		X: pub.X,
		Y: pub.Y,
	}

	buf, err := json.Marshal(key)
	if err != nil {
		t.Fatal(err)
	}

	netKey := P256PublicKey{}
	err = json.Unmarshal(buf, &netKey)
	newPub := key.Key()
	log.Printf("is equal: %v", newPub.Equal(pub))
}
