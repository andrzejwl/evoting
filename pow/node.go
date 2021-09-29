package pow

import "fmt"

type Node struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}

func (n Node) String() string {
	return fmt.Sprintf("%v:%v", n.Address, n.Port)
}
