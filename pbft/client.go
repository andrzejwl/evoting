package pbft

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// This will be the client that makes block requests.

type Request struct {
	Transactions []Transaction `json:"transactions"`
}

type Commit struct {
	From    string `json:"node-id"`
	BlockId int    `json:"block-id"`
}

type PendingRequests struct {
	requests map[int][]Commit // blockId -> slice containing all commits to a block
	nodes    []Node
}

func (pending PendingRequests) MaximumFaultyNodes() int {
	// Max faulty nodes for PBFT is floor((n-1)/3).
	// slight change compared to the same method in Blockchain struct - not counting self to n

	return int((len(pending.nodes) - 1) / 3)
}

func (pending PendingRequests) SelectPrimaryReplica() Node {
	if len(pending.nodes) == 0 {
		return Node{}
	}

	// select random node
	rand.Seed(time.Now().Unix())
	n := rand.Int() % len(pending.nodes)
	return pending.nodes[n]
}

func (pending *PendingRequests) HttpHandler(port int) {
	r := mux.NewRouter()
	r.HandleFunc("/commit", pending.ReceiveCommitInfo).Methods("POST")
	r.HandleFunc("/new-request", pending.CreateRequest).Methods("POST")

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), r))
}

func (pendingRequests *PendingRequests) ReceiveCommitInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var newCommit Commit
	decodingErr := json.NewDecoder(r.Body).Decode(&newCommit)
	if decodingErr != nil {
		http.Error(w, HttpJsonBodyPadding("incorrect request body"), http.StatusBadRequest)
		return
	}

	commits := pendingRequests.requests[newCommit.BlockId]
	commits = append(commits, newCommit)
	pendingRequests.requests[newCommit.BlockId] = commits

	f := pendingRequests.MaximumFaultyNodes()
	if len(pendingRequests.requests[newCommit.BlockId]) > f {
		delete(pendingRequests.requests, newCommit.BlockId)
		// maybe notify the web server about successful request submission
	}
}

func (pendingRequests *PendingRequests) CreateRequest(w http.ResponseWriter, r *http.Request) {
	// TODO: perhaps validate if the request is coming from a trusted party?

	w.Header().Set("Content-Type", "application/json")
	var request Request
	error := json.NewDecoder(r.Body).Decode(&request)
	if error != nil {
		http.Error(w, HttpJsonBodyPadding("incorrect request body"), http.StatusBadRequest)
		return
	}

	bodyBuffer, bufferErr := json.Marshal(request.Transactions)

	if bufferErr != nil {
		http.Error(w, HttpJsonBodyPadding("incorrect request body"), http.StatusBadRequest)
		return
	}

	// get new primary replica
	node := pendingRequests.SelectPrimaryReplica()

	response, httpErr := http.Post(fmt.Sprintf("http://%v/request", node.String()), "application/json", bytes.NewBuffer(bodyBuffer))
	if httpErr != nil {
		http.Error(w, HttpJsonBodyPadding("failed to connect to blockchain"), http.StatusBadRequest)
		return
	}

	if response.StatusCode != 200 {
		body, _ := ioutil.ReadAll(response.Body)
		http.Error(w, HttpJsonBodyPadding("blockchain error: "+string(body)), http.StatusBadRequest)
		return
	}

	var block Block
	decodingErr := json.NewDecoder(response.Body).Decode(&block)

	if decodingErr != nil {
		http.Error(w, HttpJsonBodyPadding("blockchain error: "+decodingErr.Error()), http.StatusBadRequest)
		return
	}

	fmt.Fprint(w, json.NewEncoder(w).Encode(HttpJsonBodyPadding("request submitted to the blockchain")))
}

func StartClient(httpPort int) {
	var pending PendingRequests

	fmt.Println("[CLIENT] Starting HTTP Listener")
	pending.HttpHandler(httpPort)
}
