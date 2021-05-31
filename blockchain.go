package main

import (
	"time"
)

type Blockchain struct {
	chain               []Block
	difficulty          int
	pendingTransactions []Transaction
}

func NewBlockchain(difficulty int) *Blockchain {
	initBlock := Block{timestamp: int(time.Now().Unix()), nonce: 0, transactions: []Transaction{}}
	initBlock.ProofOfWork(difficulty) // only solve the hash (modify nonce), no need to store its value

	return &Blockchain{
		chain:               []Block{initBlock},
		difficulty:          difficulty,
		pendingTransactions: []Transaction{},
	}
}

func (bc Blockchain) length() int {
	return len(bc.chain)
}

func (bc Blockchain) LastBlock() Block {
	if bc.length() == 0 {
		return Block{}
	}
	return bc.chain[bc.length()-1]
}

func (bc *Blockchain) AddTransaction(t Transaction) string {
	// first verify and return error so that the user can be notified
	error := ""
	bc.pendingTransactions = append(bc.pendingTransactions, t)

	return error
}

func (bc *Blockchain) ValidateTransactions() {
	// validate pending transactions and append them to the blockchain
	lastBlock := bc.LastBlock()
	newBlock := Block{}

	// use the commented way if validation is required

	newBlock.transactions = append(newBlock.transactions, bc.pendingTransactions...)

	// for _, t := range bc.pendingTransactions {
	// 	// validate here?
	// 	newBlock.transactions = append(newBlock.transactions, t)
	// }

	bc.pendingTransactions = []Transaction{} // clear pending transactions
	newBlock.previousBlockHash = calculateHash(lastBlock)
	newBlock.timestamp = int(time.Now().Unix())
	newBlock.ProofOfWork(bc.difficulty)
	bc.chain = append(bc.chain, newBlock)
}
