package main

import (
	"crypto/sha256"
	"fmt"
)

/*
TODO: transaction validation (has user not yet voted etc.)
subtasks:
- query to check whether a user has voted
- store globals - node addresses,
*/

type Node struct {
	address string
	port    int
}

func calculateHash(block Block) string {
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprint("%v", block)))

	return fmt.Sprintf("%x", hash.Sum(nil)) // return string representing hex formatted hash
}

func main() {
	fmt.Println("Blockchain")

	t1 := Transaction{tokenId: "qqqq-wwww-vvvv-aaaa", toId: "abc"}
	t2 := Transaction{tokenId: "qqqq-wwww-vvvv-bbbb", toId: "abc"}

	blockchain := NewBlockchain(5)

	blockchain.AddTransaction(t1)
	blockchain.AddTransaction(t2)

	// fmt.Println(block1.ProofOfWork(5))
	// fmt.Println(block1.previousBlockHash)

	fmt.Println(blockchain)
	fmt.Println("Pending transactions")
	fmt.Println(blockchain.pendingTransactions)

	fmt.Println("After validation")
	blockchain.ValidateTransactions()
	fmt.Println(blockchain)
}
