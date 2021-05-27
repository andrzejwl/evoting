package main

import (
	"crypto/sha256"
	"fmt"
	"time"
)

// TODO: block hash
// TODO: transaction validation (has user not yet voted etc.)

type Transaction struct {
	// not storing issuer ID for privacy reasons (tokenId instead)
	// not storing amount because it is always a single vote
	tokenId string
	toId    string
}

type Block struct {
	timestamp    int
	nonce        int
	transactions []Transaction
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

	block1 := Block{timestamp: int(time.Now().Unix()), nonce: 0, transactions: []Transaction{t1, t2}}

	fmt.Println(block1)
	fmt.Println(calculateHash(block1))
}
