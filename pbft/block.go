package pbft

import (
	"crypto/sha256"
	"fmt"
)

type Block struct {
	Identifier        int           `json:"id"`
	Timestamp         int           `json:"timestamp"`
	Transactions      []Transaction `json:"transactions"`
	PreviousBlockHash string        `json:"previousHash"`
}

func (b Block) AddTransaction(ta Transaction) {
	b.Transactions = append(b.Transactions, ta)
}

func calculateHash(block Block) string {
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%v", block)))

	return fmt.Sprintf("%x", hash.Sum(nil)) // return string representing hex formatted hash
}
