package main

import (
	"fmt"
)

/*
TODO: transaction validation (has user not yet voted etc.)
subtasks:
- query to check whether a user has voted (outside of blockchain)
- query to check if block exists on the chain (return index)
- store globals - node addresses,
- (?) thread responsible for achieving consensus
*/

const blockchainDifficulty int = 3

func main() {
	allowedNodeAdresses := []Node{
		Node{address: "127.0.0.1", port: 5000},
		Node{address: "127.0.0.1", port: 5001},
		Node{address: "127.0.0.1", port: 5002},
	}

	fmt.Println(allowedNodeAdresses)

	t1 := Transaction{TokenId: "qqqq-wwww-vvvv-aaaa", ToId: "abc"}
	t2 := Transaction{TokenId: "qqqq-wwww-vvvv-bbbb", ToId: "abc"}

	blockchain := NewBlockchain(blockchainDifficulty)

	blockchain.AddTransaction(t1)
	blockchain.AddTransaction(t2)
	blockchain.ValidateTransactions()

	t3 := Transaction{TokenId: "iiii-wwww-vvvv-aaaa", ToId: "abc"}
	t4 := Transaction{TokenId: "jjjj-wwww-vvvv-bbbb", ToId: "abc"}
	var bc2 = Blockchain(*blockchain)

	bc2.AddTransaction(t3)
	bc2.AddTransaction(t4)
	bc2.ValidateTransactions()

	fmt.Println("Before consensus:", blockchain.Chain)

	blockchain.Consensus(bc2)

	fmt.Println("After consensus:", blockchain.Chain)

	fmt.Println("Starting HTTP server")
	handleRequests(blockchain)
}
