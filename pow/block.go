package pow

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

type Block struct {
	Timestamp         int           `json:"timestamp"`
	Nonce             int           `json:"nonce"`
	Transactions      []Transaction `json:"transactions"`
	PreviousBlockHash string        `json:"previousHash"`
}

func calculateHash(block Block) string {
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%v", block)))

	return fmt.Sprintf("%x", hash.Sum(nil)) // return string representing hex formatted hash
}

func (b *Block) ProofOfWork(difficulty int) string {
	/*
		Calculates the PoW for the block with given difficulty.
		As a byproduct also modifies the block nonce.
	*/
	hash := calculateHash(*b)

	for !strings.HasPrefix(hash, strings.Repeat("0", difficulty)) {
		b.Nonce += 1
		hash = calculateHash(*b)
	}

	return hash
}

func (b Block) AddTransaction(ta Transaction) {
	valid, error := validateTransaction(ta)

	if !valid {
		fmt.Println(error)
		return
	}

	b.Transactions = append(b.Transactions, ta)

	// reset nonce when block changes
	b.Nonce = 0
}

func InitBlock() {

}
