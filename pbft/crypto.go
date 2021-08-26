package pbft

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
)

func GenerateSigningKeyPair() (rsa.PrivateKey, *rsa.PublicKey) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	return *privateKey, &privateKey.PublicKey
}

func SignData(message []byte, priv *rsa.PrivateKey) ([]byte, error) {
	hash := sha256.Sum256(message)

	return rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hash[:])
}

func VerifySignature(pub *rsa.PublicKey, signed []byte, message string) error {
	// sigHash := sha256.Sum256(signed)
	msgHash := sha256.Sum256([]byte(message))
	// unhexlify string

	signUnhexed := make([]byte, hex.DecodedLen(len(signed)))
	_, err := hex.Decode(signUnhexed, signed)

	if err != nil {
		return err
	}

	return rsa.VerifyPKCS1v15(pub, crypto.SHA256, msgHash[:], signUnhexed)
}
