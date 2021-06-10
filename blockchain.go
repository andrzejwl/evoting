package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Blockchain struct {
	Chain               []Block       `json:"chain"`
	Difficulty          int           `json:"difficulty"`
	pendingTransactions []Transaction // no need to export this field
	blockHashes         []string      // maintained so that hashes are not calculated all the time
}

func NewBlockchain(difficulty int) *Blockchain {
	initBlock := Block{Timestamp: int(time.Now().Unix()), Nonce: 0, Transactions: []Transaction{}}
	initBlock.ProofOfWork(difficulty) // only solve the hash (modify nonce), no need to store its value

	return &Blockchain{
		Chain:               []Block{initBlock},
		Difficulty:          difficulty,
		pendingTransactions: []Transaction{},
	}
}

func (bc Blockchain) length() int {
	return len(bc.Chain)
}

func (bc Blockchain) LastBlock() Block {
	if bc.length() == 0 {
		return Block{}
	}
	return bc.Chain[bc.length()-1]
}

func (bc Blockchain) GenesisBlock() Block {
	if bc.length() == 0 {
		return Block{}
	}
	return bc.Chain[0]
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

	newBlock.Transactions = append(newBlock.Transactions, bc.pendingTransactions...)

	// for _, t := range bc.pendingTransactions {
	// 	// validate here?
	// 	newBlock.transactions = append(newBlock.transactions, t)
	// }

	bc.pendingTransactions = []Transaction{} // clear pending transactions
	newBlock.PreviousBlockHash = calculateHash(lastBlock)
	newBlock.Timestamp = int(time.Now().Unix())
	newBlock.ProofOfWork(bc.Difficulty)
	bc.Chain = append(bc.Chain, newBlock)
	bc.blockHashes = append(bc.blockHashes, calculateHash(newBlock))
}

func (bc *Blockchain) HttpGetChain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, json.NewEncoder(w).Encode(bc))
}

func (bc *Blockchain) HttpCreateTransaction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var t Transaction

	error := json.NewDecoder(r.Body).Decode(&t)
	if error != nil {
		http.Error(w, error.Error(), http.StatusBadRequest)
		return
	}

	bc.AddTransaction(t)
	bc.ValidateTransactions()
	fmt.Fprint(w, json.NewEncoder(w).Encode(bc))
}

func (bc Blockchain) IsValid() bool {
	for _, block := range bc.Chain {
		if !strings.HasPrefix(calculateHash(block), strings.Repeat("0", bc.Difficulty)) {
			return false
		}
	}
	return true
}

func (bc *Blockchain) Consensus(outsideChain Blockchain) {
	/*
		Verifies if outside chain is correct, if so, appends new blocks
	*/
	if !outsideChain.IsValid() || bc.length() >= outsideChain.length() {
		return
	}

	if calculateHash(bc.GenesisBlock()) == calculateHash(outsideChain.GenesisBlock()) {
		// genesis blocks are the same and outside chain is valid => bc is a subchain of the outside chain
		bc.Chain = outsideChain.Chain
	}
}
