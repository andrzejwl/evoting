package main

/*
Simple server that keeps track of all nodes and their type (blockchain/client).
*/

import (
	"crypto/rsa"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Node struct {
	Address    string         `json:"address"`
	Port       int            `json:"port"`
	Type       string         `json:"node-type"`
	Identifier string         `json:"node-id"`
	PublicKey  *rsa.PublicKey `json:"public-key"`
}

func (n Node) String() string {
	return fmt.Sprintf("%v:%v", n.Address, n.Port)
}

type NodeDiscovery struct {
	BlockchainNodes []Node
	ClientNodes     []Node
	VotingParties   []VotingParty
}

type VotingParty struct {
	Identifier string `json:"id"`
}

func NewDiscovery() *NodeDiscovery {
	var nd NodeDiscovery
	return &nd
}

func (nd *NodeDiscovery) GetPartyById(id string) *VotingParty {
	for _, p := range nd.VotingParties {
		if p.Identifier == id {
			return &p
		}
	}
	return nil
}

func (nd *NodeDiscovery) HttpGetAllNodes(w http.ResponseWriter, r *http.Request) {
	var all []Node
	all = append(all, nd.BlockchainNodes...)
	all = append(all, nd.ClientNodes...)

	json.NewEncoder(w).Encode(all)
}

func (nd *NodeDiscovery) HttpGetBlockchain(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(nd.BlockchainNodes)
}

func (nd *NodeDiscovery) HttpGetClients(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(nd.ClientNodes)
}

func (nd *NodeDiscovery) HttpGetVotingParties(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(nd.VotingParties)
}

func (nd *NodeDiscovery) HttpRegisterParty(w http.ResponseWriter, r *http.Request) {
	var newParty VotingParty
	decodingErr := json.NewDecoder(r.Body).Decode(&newParty)

	if decodingErr != nil {
		http.Error(w, "error parsing request body", http.StatusBadRequest)
		return
	}

	if nd.GetPartyById(newParty.Identifier) == nil {
		nd.VotingParties = append(nd.VotingParties, newParty)
	}

	json.NewEncoder(w).Encode(nd.VotingParties)
}

func ContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func (nd *NodeDiscovery) HttpRegisterNode(w http.ResponseWriter, r *http.Request) {
	var newNode Node
	decodingErr := json.NewDecoder(r.Body).Decode(&newNode)
	if decodingErr != nil {
		http.Error(w, "{\"detail\": \"incorrect request body\"}", http.StatusBadRequest)
		return
	}

	if newNode.Type == "blockchain" {
		nd.BlockchainNodes = append(nd.BlockchainNodes, newNode)
	} else if newNode.Type == "client" {
		nd.ClientNodes = append(nd.ClientNodes, newNode)
	} else {
		http.Error(w, "{\"detail\": \"incorrect node type\"}", http.StatusBadRequest)
		return
	}
	fmt.Println("[INFO] New peer registered", newNode)
	json.NewEncoder(w).Encode(`{"detail":"ok"}`)
}

func HandleRequests(port int, nd *NodeDiscovery) {
	r := mux.NewRouter()
	r.Use(ContentTypeMiddleware)

	r.HandleFunc("/get-all", nd.HttpGetAllNodes).Methods("GET")
	r.HandleFunc("/get-blockchain", nd.HttpGetBlockchain).Methods("GET")
	r.HandleFunc("/get-clients", nd.HttpGetClients).Methods("GET")
	r.HandleFunc("/register", nd.HttpRegisterNode).Methods("POST")

	r.HandleFunc("/get-parties", nd.HttpGetVotingParties).Methods("GET")
	r.HandleFunc("/register-party", nd.HttpRegisterParty).Methods("POST")

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), r))
}

func main() {
	nd := NewDiscovery()
	portPtr := flag.Int("port", 9999, "HTTP listener port")

	fmt.Println("[Node Discovery] Starting HTTP Listener on port", *portPtr)
	HandleRequests(*portPtr, nd)
}
