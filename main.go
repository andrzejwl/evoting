package main

import (
	"evoting/pbft"
	"evoting/pow"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
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
		var hostname string
		var port int
		if os.Getenv("DOCKER") == "1" {
			hostname = os.Getenv("HOSTNAME")
			port, _ = strconv.Atoi(os.Getenv("PORT"))
		} else {
			hostname = "127.0.0.1"
			port = *portPtr
		}

		blockchain := pow.NewBlockchain(blockchainDifficulty, hostname, port)

		if !*rootPtr {
			// not the "root" node - fetch existing chain from other peers
			fmt.Println("Fetching chain from peers")
			if os.Getenv("DOCKER") == "1" {
				peerHostname := os.Getenv("PEER_HOSTNAME")
				peerPort, _ := strconv.Atoi(os.Getenv("PEER_PORT"))
				blockchain.Peers = append(blockchain.Peers,
					pow.Node{Address: peerHostname, Port: peerPort},
				)
			} else {
				blockchain.Peers = append(blockchain.Peers,
					pow.Node{Address: "127.0.0.1", Port: *peerPortPtr},
				)
			}

			blockchain.Update(true)
		}
		// else {
		// 	t1 := pow.Transaction{TokenId: "qqqq-wwww-vvvv-aaaa", ToId: "abc"}
		// 	t2 := pow.Transaction{TokenId: RandomString(16), ToId: RandomString(5)}

		// 	blockchain.AddTransaction(t1)
		// 	blockchain.AddTransaction(t2)
		// 	blockchain.ValidateTransactions()
		// }
		fmt.Println("Starting HTTP server on port", blockchain.Self.Port)
		blockchain.PropagateSelf()
		pow.HandleRequests(blockchain)
	} else if *consensusPtr == "pbft" {
		// Practical Byzantine Fault Tolerance
		blockchain := pbft.NewBlockchain(*portPtr)
		fmt.Println("Starting HTTP server on port", *portPtr)
		pbft.HandleRequests(*portPtr, blockchain)
	}
}
