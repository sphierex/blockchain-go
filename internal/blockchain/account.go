package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"

	"golang.org/x/crypto/ripemd160"

	"github.com/sphierex/blockchain-go/pkg/base58"
)

const (
	accountVersion     = byte(0x00)
	accountChecksumLen = 4
)

// Account stores private and public keys.
type Account struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// NewAccount creates and returns a Account
func NewAccount() *Account {
	private, public := newKeyPair()

	account := Account{
		PrivateKey: private,
		PublicKey:  public,
	}

	return &account
}

// Address returns account address.
func (a *Account) Address() []byte {
	pubKeyHash := HashPubKey(a.PublicKey)

	versionPayload := append([]byte{accountVersion}, pubKeyHash...)
	payload := append(versionPayload, checksum(versionPayload)...)

	return base58.Encode(payload)
}

func (a *Account) String() string {
	return fmt.Sprintf("%s", a.Address())
}

// HashPubKey hashes public key
func HashPubKey(pubKey []byte) []byte {
	pubHash := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, _ = RIPEMD160Hasher.Write(pubHash[:])
	pubRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return pubRIPEMD160
}

// ValidateAddress check if address is valid.
func ValidateAddress(address string) bool {
	if len(address) != 34 {
		return false
	}

	pubKeyHash := base58.Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-accountChecksumLen:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-accountChecksumLen]
	targetChecksum := checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0
}

// Checksum generates a checksum for a public key
func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:accountChecksumLen]
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, _ := ecdsa.GenerateKey(curve, rand.Reader)
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, pubKey
}
