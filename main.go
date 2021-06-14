package main

import (
	"flag"
	"fmt"
	"math/rand"
)

/*
TODO: transaction validation (has user not yet voted etc.)
subtasks:
- propagate valid transactions before block is appended (so that other peers do not append new blocks)
- query to check whether a user has voted (outside of blockchain)
- store globals - node addresses,
- (?) thread responsible for achieving consensus
*/

const blockchainDifficulty int = 3

func RandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func main() {
	// cli params
	portPtr := flag.Int("port", 5000, "HTTP server port")
	peerPtr := flag.String("peer", "127.0.0.1:5001", "Peer address")

	flag.Parse()
	fmt.Println(*portPtr, *peerPtr)

	t1 := Transaction{TokenId: "qqqq-wwww-vvvv-aaaa", ToId: "abc"}
	t2 := Transaction{TokenId: RandomString(16), ToId: RandomString(5)}

	blockchain := NewBlockchain(blockchainDifficulty)

	blockchain.peers = append(blockchain.peers,
		Node{address: "127.0.0.1", port: 5000},
		Node{address: "127.0.0.1", port: 5001},
		Node{address: "127.0.0.1", port: 5002},
	)

	blockchain.AddTransaction(t1)
	blockchain.AddTransaction(t2)
	blockchain.ValidateTransactions()

	// t3 := Transaction{TokenId: "iiii-wwww-vvvv-aaaa", ToId: "abc"}
	// t4 := Transaction{TokenId: "jjjj-wwww-vvvv-bbbb", ToId: "abc"}
	// var bc2 = Blockchain(*blockchain)

	// bc2.AddTransaction(t3)
	// bc2.AddTransaction(t4)
	// bc2.ValidateTransactions()

	// fmt.Println("Before consensus:", blockchain.Chain)

	// blockchain.Consensus(bc2)

	// fmt.Println("After consensus:", blockchain.Chain)

	fmt.Println("Starting HTTP server on port", *portPtr)
	handleRequests(*portPtr, blockchain)
}
