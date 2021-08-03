package main

import (
	"evoting/pbft"
	"evoting/pow"
	"flag"
	"fmt"
	"math/rand"
)

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
	clientPtr := flag.Bool("client_mode", false, "run client mode (middleman between web app and blockchain)")
	portPtr := flag.Int("port", 5000, "HTTP server port")
	rootPtr := flag.Bool("root", false, "Is node the root node - initialize a new chain")
	peerPortPtr := flag.Int("peer", 5001, "Localhost peer port flag")
	consensusPtr := flag.String("consensus", "pow", "Consensus mechanism: pow / poa / pbft")

	flag.Parse()

	if *clientPtr {
		pbft.StartClient(*portPtr)
		return
	}

	if *consensusPtr == "pow" {
		// Proof of Work
		blockchain := pow.NewBlockchain(blockchainDifficulty)

		if !*rootPtr {
			// not the "root" node - fetch existing chain from other peers
			fmt.Println("Fetching chain from peers")
			blockchain.Peers = append(blockchain.Peers,
				pow.Node{Address: "127.0.0.1", Port: *peerPortPtr},
			)
			blockchain.Update(true)
		} else {
			t1 := pow.Transaction{TokenId: "qqqq-wwww-vvvv-aaaa", ToId: "abc"}
			t2 := pow.Transaction{TokenId: RandomString(16), ToId: RandomString(5)}

			blockchain.AddTransaction(t1)
			blockchain.AddTransaction(t2)
			blockchain.ValidateTransactions()
		}

		fmt.Println("Starting HTTP server on port", *portPtr)
		pow.HandleRequests(*portPtr, blockchain)
	} else if *consensusPtr == "poa" {
		// Proof of Authority
		fmt.Println("PoA")
	} else if *consensusPtr == "pbft" {
		// Practical Byzantine Fault Tolerance
		fmt.Println("PBFT")
		blockchain := pbft.NewBlockchain()
		fmt.Println("Starting HTTP server on port", *portPtr)
		pbft.HandleRequests(*portPtr, blockchain)
	}
}
