package main

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"fmt"
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

type Transaction struct {
	TokenId string `json:"Token"`
	ToId    string `json:"ToId"`
}

type VotingParty struct {
	Identifier string `json:"id"`
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

func JsonBodyPadding(message string) string {
	body := fmt.Sprintf("{\"detail\": \"%v\"}", message)
	return body
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

	ndAddr := os.Getenv("ND_ADDR")

	if ndAddr == "" {
		ndAddr = "127.0.0.1:9999" // debug address
	}

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
	fmt.Fprint(w, json.NewEncoder(w).Encode(JsonBodyPadding("request submitted to the blockchain")))
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

	ndAddr := os.Getenv("ND_ADDR")

	if ndAddr == "" {
		ndAddr = "127.0.0.1:9999" // debug address
	}

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
	fmt.Fprint(w, json.NewEncoder(w).Encode(JsonBodyPadding("party registered")))
}

func Handler(port int) {
	r := mux.NewRouter()

	r.HandleFunc("/add-data", HttpAddData).Methods("POST")
	r.HandleFunc("/add-voting-party", HttpAddVotingParty).Methods("POST")

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), r))
}

func main() {
	port := 1234
	fmt.Println("[INFO] starting HTTP server on port", port)
	Handler(port)
}
