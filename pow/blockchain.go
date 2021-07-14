package pow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const DEBUG_MODE bool = false // TODO: set this dynamically in main.go

type Blockchain struct {
	Chain               []Block       `json:"chain"`
	Difficulty          int           `json:"difficulty"`
	pendingTransactions []Transaction // no need to export this field
	blockHashes         []string      // maintained so that hashes are not calculated all the time
	Peers               []Node        `json:"-"` // right now not shared but could be used to propagate new peers
}

func NewBlockchain(difficulty int) *Blockchain {
	var initBlock Block
	if DEBUG_MODE {
		initBlock = Block{Timestamp: 0, Nonce: 0, Transactions: []Transaction{}}
	} else {
		initBlock = Block{Timestamp: int(time.Now().Unix()), Nonce: 0, Transactions: []Transaction{}}
	}
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
	bc.Update(false)

	newBlock := Block{}
	lastBlock := bc.LastBlock()

	// use the commented way if validation is required

	newBlock.Transactions = append(newBlock.Transactions, bc.pendingTransactions...)

	// for _, t := range bc.pendingTransactions {
	// 	// validate here?
	// 	newBlock.transactions = append(newBlock.transactions, t)
	// }

	bc.pendingTransactions = []Transaction{} // clear pending transactions
	newBlock.PreviousBlockHash = calculateHash(lastBlock)

	if DEBUG_MODE {
		newBlock.Timestamp = 0
	} else {
		newBlock.Timestamp = int(time.Now().Unix())
	}
	newBlock.ProofOfWork(bc.Difficulty)
	bc.Chain = append(bc.Chain, newBlock)
	bc.blockHashes = append(bc.blockHashes, calculateHash(newBlock))
	bc.PropagateChain()
}

func (bc *Blockchain) HttpGetChain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Println("GET /chain Request from:", r.RemoteAddr)
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
	// in the future ValidateTransactions should not be called separately for each transaction
	// might need to add a database to store the transactions in
	bc.ValidateTransactions()
	fmt.Fprint(w, json.NewEncoder(w).Encode(bc))
}

func (bc *Blockchain) HttpUpdate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	newChain, err := ReconstructBlockchain(r.Body)

	if err != "" {
		http.Error(w, err, http.StatusBadRequest)
		return
	}
	bc.Consensus(newChain)
	fmt.Fprint(w, json.NewEncoder(w).Encode("{\"detail\":\"ok\"}"))
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

func ReconstructBlockchain(r io.ReadCloser) (Blockchain, string) {
	var bc Blockchain
	decodingErr := json.NewDecoder(r).Decode(&bc)

	if decodingErr != nil {
		return Blockchain{}, decodingErr.Error()
	}

	// fill this as blockHashes are not transmitted over network
	bc.blockHashes = nil // assert blockHashes is an empty slice
	for _, block := range bc.Chain {
		bc.blockHashes = append(bc.blockHashes, calculateHash(block))
	}

	return bc, ""
}

func (bc *Blockchain) Update(initialize bool) {
	// TODO: check if pending transactions have not already been validated and appended by other nodes
	for _, peer := range bc.Peers {
		resp, err := http.Get(fmt.Sprintf("http://%v/chain/", peer.String()))
		if err != nil {
			fmt.Println("[ERR]", err, "Peer chain check failed, skipping peer", peer)
			continue
		}

		peerBc, reconstructErr := ReconstructBlockchain(resp.Body)
		if reconstructErr != "" {
			fmt.Println("[ERR] Peer chain check failed, skipping peer", peer)
			continue
		}
		if initialize {
			bc.Chain = peerBc.Chain
			bc.Difficulty = peerBc.Difficulty
		} else {
			bc.Consensus(peerBc)
		}
	}
}

func (bc *Blockchain) HttpTriggerUpdate(w http.ResponseWriter, r *http.Request) {
	/*
		Temporary function used for debugging and testing.
	*/
	bc.Update(false)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, "{\"detail\": \"ok\"}")
}

func (bc Blockchain) PropagateChain() {
	bcBuffer, jsonErr := json.Marshal(bc)

	if jsonErr != nil {
		fmt.Println("[ERR] Chain encoding error", jsonErr.Error())
		return
	}

	for _, peer := range bc.Peers {
		fmt.Println("Propagating to", peer)
		resp, err := http.Post(fmt.Sprintf("http://%v/update", peer.String()), "application/json", bytes.NewBuffer(bcBuffer))

		if err != nil {
			fmt.Println("[ERR] Propagation failed to", peer, ", error:", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, respErr := ioutil.ReadAll(resp.Body)
			if respErr != nil {
				fmt.Println("[ERR]", respErr.Error())
			}
			fmt.Println("[ERR] Response from", peer, ": status code:", resp.StatusCode, ", response:", string(body))
		}
	}
}