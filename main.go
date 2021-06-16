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

const DEBUG_MODE bool = false

func main() {
	// cli params
	portPtr := flag.Int("port", 5000, "HTTP server port")
	rootPtr := flag.Bool("root", false, "Peer address")
	typePtr := flag.String("type", "A", "[Debug]: A/B")
	peerPortPtr := flag.Int("peer", 5001, "Localhost peer port flag")

	flag.Parse()

	blockchain := NewBlockchain(blockchainDifficulty)

	if !*rootPtr {
		// not the "root" node - fetch existing chain from other peers
		fmt.Println("Fetching chain from peers")
		blockchain.peers = append(blockchain.peers,
			Node{address: "127.0.0.1", port: *peerPortPtr},
		)
		blockchain.Update(true)
	} else {
		t1 := Transaction{TokenId: "qqqq-wwww-vvvv-aaaa", ToId: "abc"}
		t2 := Transaction{TokenId: RandomString(16), ToId: RandomString(5)}

		blockchain.AddTransaction(t1)
		blockchain.AddTransaction(t2)
		blockchain.ValidateTransactions()
	}

	if *typePtr == "B" {
		t3 := Transaction{TokenId: "iiii-wwww-vvvv-aaaa", ToId: "abc"}
		t4 := Transaction{TokenId: "jjjj-wwww-vvvv-bbbb", ToId: "abc"}

		blockchain.AddTransaction(t3)
		blockchain.AddTransaction(t4)
		blockchain.ValidateTransactions()
	}
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
