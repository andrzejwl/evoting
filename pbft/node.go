package pbft

import "fmt"

type Node struct {
	Address    string `json:"address"`
	Port       int    `json:"port"`
	Identifier string `json:"node-id"`
	Type       string `json:"node-type"`
	PublicKey  string `json:"public-key"`
	privateKey string // do not export private key
}

func (n Node) String() string {
	return fmt.Sprintf("%v:%v", n.Address, n.Port)
}
