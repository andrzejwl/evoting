package pow

import "fmt"

type Node struct {
	Address string
	Port    int
}

func (n Node) String() string {
	return fmt.Sprintf("%v:%v", n.Address, n.Port)
}
