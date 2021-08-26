package pbft

import (
	"crypto/rsa"
	"fmt"
)

type Node struct {
	Address    string          `json:"address"`
	Port       int             `json:"port"`
	Identifier string          `json:"node-id"`
	Type       string          `json:"node-type"`
	PublicKey  *rsa.PublicKey  `json:"public-key"`
	privateKey *rsa.PrivateKey // do not export private key
}

func (n Node) String() string {
	return fmt.Sprintf("%v:%v", n.Address, n.Port)
}
