package pbft

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

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// This will be the client that makes block requests.

type Request struct {
	Transactions []Transaction `json:"transactions"`
	Client       Node          `json:"requesting-client"`
}

type Commit struct {
	From    string `json:"node-id"`
	BlockId int    `json:"block-id"`
}

type PendingRequests struct {
	requests         map[int][]Commit // blockId -> slice containing all commits to a block
	nodes            []Node
	discoveryAddress string
	self             Node
}

func (pending PendingRequests) MaximumFaultyNodes() int {
	// Max faulty nodes for PBFT is floor((n-1)/3).
	// slight change compared to the same method in Blockchain struct - not counting self to n

	return int((len(pending.nodes) - 1) / 3)
}

func (pending *PendingRequests) RegisterNode(addr string, port int, idnt string, pubKey *rsa.PublicKey, privKey *rsa.PrivateKey) {
	self := Node{Address: addr, Port: port, Identifier: idnt, Type: "client", PublicKey: pubKey, privateKey: privKey}
	pending.self = self

	messageBuffer, _ := json.Marshal(pending.self)
	resp, err := http.Post(fmt.Sprintf("http://%v/register", pending.discoveryAddress), "application/json", bytes.NewBuffer(messageBuffer))

	if err != nil || resp.StatusCode != 200 {
		fmt.Println("[ERROR] failed to register node at node discovery service")
		if err != nil {
			fmt.Println("details:", err.Error())
		} else {
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			fmt.Println("status code:", resp.StatusCode, ", discovery response:", string(bodyBytes))
		}
	} else {
		fmt.Println("[INFO] Node registered")
	}
}

func (pending *PendingRequests) RefreshNodes() {
	var newNodes []Node
	resp, err := http.Get(fmt.Sprintf("http://%v/get-blockchain", pending.discoveryAddress))

	if err != nil {
		fmt.Println("[CLIENT] failed to refresh nodes")
		return
	}

	decodingErr := json.NewDecoder(resp.Body).Decode(&newNodes)

	if decodingErr != nil {
		fmt.Println("[CLIENT] node discovery's response is ambiguous")
		return
	}

	pending.nodes = newNodes
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
	request.Client = pendingRequests.self

	error := json.NewDecoder(r.Body).Decode(&request)
	if error != nil {
		http.Error(w, HttpJsonBodyPadding("incorrect request body"), http.StatusBadRequest)
		return
	}

	bodyBuffer, bufferErr := json.Marshal(request)

	if bufferErr != nil {
		http.Error(w, HttpJsonBodyPadding("incorrect request body"), http.StatusBadRequest)
		return
	}

	// get new primary replica
	pendingRequests.RefreshNodes()
	node := pendingRequests.SelectPrimaryReplica()
	fmt.Println("[DEBUG] primary replica:", node.Identifier)
	fmt.Println("[DEBUG] rqeuest:", request)

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
	pending.discoveryAddress = os.Getenv("DISCOVERY_ADDR")

	priv, pub := GenerateSigningKeyPair()
	pending.RegisterNode(os.Getenv("HOSTNAME"), httpPort, uuid.NewString(), pub, &priv)

	fmt.Println("[CLIENT] Starting HTTP Listener")
	pending.HttpHandler(httpPort)
}
