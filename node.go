package main

import "fmt"

type Node struct {
	address string
	port    int
}

func (n Node) String() string {
	return fmt.Sprintf("%v:%v", n.address, n.port)
}
