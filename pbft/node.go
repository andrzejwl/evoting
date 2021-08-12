package pbft

import "fmt"

type Node struct {
	Address    string `json:"address"`
	Port       int    `json:"port"`
	Identifier string `json:"node-id"`
	Type       string `json:"node-type"`
}

func (n Node) String() string {
	return fmt.Sprintf("%v:%v", n.Address, n.Port)
}
