package internal

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"

	"github.com/sphierex/blockchain-go/pkg"
)

const (
	version            = byte(0x00)
	addressChecksumLen = 4
)

type Account struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func NewAccount() *Account {
	private, public := newKeyPair()
	wallet := Account{private, public}

	return &wallet
}

func (a *Account) GetAddress() []byte {
	pubKeyHash := HashPubKey(a.PublicKey)

	versionPayload := append([]byte{version}, pubKeyHash...)
	checksum := checksum(versionPayload)

	fullPayload := append(versionPayload, checksum...)
	address := pkg.Base58Encode(fullPayload)

	return address
}

func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, _ = RIPEMD160Hasher.Write(publicSHA256[:])
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160
}

func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen]
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {
	private, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, pubKey
}
