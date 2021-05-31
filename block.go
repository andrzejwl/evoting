package main

import (
	"fmt"
	"strings"
)

type Block struct {
	timestamp         int
	nonce             int
	transactions      []Transaction
	previousBlockHash string // chain
}

func (b Block) ProofOfWork(difficulty int) string {
	/*
		Calculates the PoW for the block with given difficulty.
		As a byproduct also modifies the block nonce.
	*/
	hash := calculateHash(b)

	for !strings.HasPrefix(hash, strings.Repeat("0", difficulty)) {
		hash = calculateHash(b)
		b.nonce += 1
	}

	return hash
}

func (b Block) AddTransaction(ta Transaction) {
	valid, error := validateTransaction(ta)

	if !valid {
		fmt.Println(error)
		return
	}

	b.transactions = append(b.transactions, ta)

	// reset nonce when block changes
	b.nonce = 0
}
