package main

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

type Node struct {
	Address    string         `json:"address"`
	Port       int            `json:"port"`
	Type       string         `json:"node-type"`
	Identifier string         `json:"node-id"`
	PublicKey  *rsa.PublicKey `json:"public-key"`
}

type Blockchain struct {
	Chain      []Block `json:"chain"`
	Identifier string  `json:"node-id"`
}

type Block struct {
	Identifier        int           `json:"id"`
	Timestamp         int           `json:"timestamp"`
	Transactions      []Transaction `json:"transactions"`
	PreviousBlockHash string        `json:"previousHash"`
}

type Transaction struct {
	TokenId string `json:"Token"`
	ToId    string `json:"ToId"`
}

type VotingParty struct {
	Identifier string `json:"id"`
}

type Results struct {
	TotalVotes int            `json:"total-votes"`
	Votes      map[string]int `json:"results"`
}

func (n Node) String() string {
	return fmt.Sprintf("%v:%v", n.Address, n.Port)
}

func RandomNode(nodes []Node) Node {
	rand.Seed(time.Now().Unix())
	n := rand.Int() % len(nodes)
	return nodes[n]
}

func ClientNodes(ndAddr string) []Node {
	var clients []Node
	req, err := http.Get(fmt.Sprintf("http://%v/get-clients", ndAddr))
	if err != nil {
		fmt.Println("[ERROR] cant connect to node discovery")
		return nil
	}

	decodingErr := json.NewDecoder(req.Body).Decode(&clients)
	if decodingErr != nil {
		fmt.Println("[ERROR] cant parse node discovery output")
		return nil
	}

	return clients
}

func BlockchainNodes(ndAddr string) []Node {
	var nodes []Node
	req, err := http.Get(fmt.Sprintf("http://%v/get-blockchain", ndAddr))
	if err != nil {
		fmt.Println("[ERROR] cant connect to node discovery")
		return nil
	}

	decodingErr := json.NewDecoder(req.Body).Decode(&nodes)
	if decodingErr != nil {
		fmt.Println("[ERROR] cant parse node discovery output")
		return nil
	}

	return nodes
}

func ReconstructBlockchain(r io.ReadCloser) (Blockchain, error) {
	var bc Blockchain
	decodingErr := json.NewDecoder(r).Decode(&bc)

	if decodingErr != nil {
		return Blockchain{}, decodingErr
	}

	return bc, nil
}

func GetBlockchainFromNode(n Node) (Blockchain, error) {
	resp, err := http.Get(fmt.Sprintf("http://%v/chain", n))
	if err != nil {
		fmt.Println("[ERROR] could not connect to peer", err.Error())
		return Blockchain{}, err
	}

	bc, decodingErr := ReconstructBlockchain(resp.Body)
	if decodingErr != nil {
		fmt.Println("[ERROR] could not parse blockchain data", decodingErr.Error())
		return Blockchain{}, decodingErr
	}

	return bc, nil
}

func Statistics(bc Blockchain) Results {
	var res Results
	res.Votes = make(map[string]int)

	for _, block := range bc.Chain {
		res.TotalVotes += len(block.Transactions)
		for _, t := range block.Transactions {
			res.Votes[t.ToId] += 1
		}
	}

	return res
}

func BlockchainFromRandomNode() (Blockchain, error) {
	ndAddr := NodeDiscoveryAddress()
	nodes := BlockchainNodes(ndAddr)
	n := RandomNode(nodes)

	bc, err := GetBlockchainFromNode(n)
	if err != nil {
		return Blockchain{}, err
	}
	return bc, nil
}

func StatisticsFromRandomNode() (Results, error) {
	bc, err := BlockchainFromRandomNode()

	if err != nil {
		return Results{}, err
	}

	results := Statistics(bc)
	return results, nil
}

func TransactionByToken(tokenId string) (Transaction, error) {
	bc, err := BlockchainFromRandomNode()

	if err != nil {
		return Transaction{}, err
	}

	for _, block := range bc.Chain {
		for _, t := range block.Transactions {
			if t.TokenId == tokenId {
				return t, nil
			}
		}
	}

	return Transaction{}, errors.New("transaction of given id not found")
}

func JsonBodyPadding(message string) string {
	body := fmt.Sprintf(`{"detail": "%v"}`, message)
	return body
}

func NodeDiscoveryAddress() string {
	ndAddr := os.Getenv("ND_ADDR")

	if ndAddr == "" {
		ndAddr = "127.0.0.1:9999" // debug address
	}

	return ndAddr
}

func HttpAddData(w http.ResponseWriter, r *http.Request) {
	/*
		create new transactions (insert new votes)
	*/

	var transactions []Transaction

	decodingErr := json.NewDecoder(r.Body).Decode(&transactions)

	if decodingErr != nil {
		http.Error(w, "request cant be parsed", http.StatusBadRequest)
		return
	}

	ndAddr := NodeDiscoveryAddress()

	clientNodes := ClientNodes(ndAddr)

	if len(clientNodes) == 0 {
		fmt.Println("[ERROR] no client nodes found")
		http.Error(w, "cant connect to blockchain", http.StatusInternalServerError)
		return
	}

	client := RandomNode(clientNodes)

	transactionsJson, _ := json.Marshal(transactions)
	reqbody := fmt.Sprintf("{\"transactions\":%v}", string(transactionsJson))
	response, error := http.Post(fmt.Sprintf("http://%v/new-request", client), "application/json", bytes.NewBuffer([]byte(reqbody)))

	if error != nil {
		fmt.Println("[ERROR] can't connect to blockchain client:", client, error.Error())
		http.Error(w, "an error occured when making a request to blockchain client", http.StatusInternalServerError)
	}

	if response.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(response.Body)
		fmt.Println("[ERROR] blockchain client error - response status code", response.StatusCode, ", body:", bytes.NewBuffer(body))
		http.Error(w, "blockchain failed to add transaction data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(JsonBodyPadding("request submitted to the blockchain"))
}

func HttpAddVotingParty(w http.ResponseWriter, r *http.Request) {
	/*
		register a new electable party at node registry
	*/
	var newParty VotingParty

	decodingErr := json.NewDecoder(r.Body).Decode(&newParty)

	if decodingErr != nil {
		http.Error(w, "incorrect request body", http.StatusBadRequest)
		return
	}

	ndAddr := NodeDiscoveryAddress()

	msgBuffer, _ := json.Marshal(newParty)
	resp, err := http.Post(fmt.Sprintf("http://%v/register-party", ndAddr), "application/json", bytes.NewBuffer(msgBuffer))

	if err != nil || resp.StatusCode != http.StatusOK {
		if err != nil {
			fmt.Println("[ERROR]", err.Error())
		} else {
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			fmt.Println("[ERROR] status code:", resp.StatusCode, ", response body:", string(bodyBytes))
		}
		http.Error(w, "failed to submit request to node discovery", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(JsonBodyPadding("party registered"))
}

func HttpGetStatistics(w http.ResponseWriter, r *http.Request) {
	stats, err := StatisticsFromRandomNode()

	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		fmt.Println("[ERROR] failed to fetch stats", err.Error())
		http.Error(w, JsonBodyPadding("failed to fetch statistics"), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(stats)
}

func HttpVerifyByToken(w http.ResponseWriter, r *http.Request) {
	/*
		if one wants to verify that their vote was submitted correctly
	*/

	var tToken Transaction // only parse tokenId
	decodingErr := json.NewDecoder(r.Body).Decode(&tToken)

	if decodingErr != nil {
		http.Error(w, JsonBodyPadding("incorrect request body"), http.StatusBadRequest)
		return
	}

	transaction, err := TransactionByToken(tToken.TokenId)
	if err != nil {
		http.Error(w, JsonBodyPadding("token not found"), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transaction)
}

func Handler(port int) {
	r := mux.NewRouter()

	r.HandleFunc("/statistics", HttpGetStatistics).Methods("GET")

	r.HandleFunc("/add-data", HttpAddData).Methods("POST")
	r.HandleFunc("/add-voting-party", HttpAddVotingParty).Methods("POST")
	r.HandleFunc("/verify", HttpVerifyByToken).Methods("POST")

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), r))
}

func main() {
	port := 1234
	fmt.Println("[INFO] starting HTTP server on port", port)
	Handler(port)
}
