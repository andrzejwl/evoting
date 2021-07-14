package pbft

import "fmt"

type Node struct {
	Address    string
	Port       int
	Identifier string
}

func (n Node) String() string {
	return fmt.Sprintf("%v:%v", n.Address, n.Port)
}
