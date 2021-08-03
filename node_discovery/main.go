package main

/*
Simple server that keeps track of all nodes and their type (blockchain/client).
*/

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Node struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
	Type    string `json:"node-type"`
}

func (n Node) String() string {
	return fmt.Sprintf("%v:%v", n.Address, n.Port)
}

type NodeDiscovery struct {
	BlockchainNodes []Node
	ClientNodes     []Node
}

func (nd NodeDiscovery) HttpGetAllNodes(w http.ResponseWriter, r *http.Request) {
	var all []Node
	all = append(all, nd.BlockchainNodes...)
	all = append(all, nd.ClientNodes...)
	fmt.Fprint(w, json.NewEncoder(w).Encode(all))
}

func (nd NodeDiscovery) HttpGetBlockchain(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, json.NewEncoder(w).Encode(nd.BlockchainNodes))
}

func (nd NodeDiscovery) HttpGetClients(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, json.NewEncoder(w).Encode(nd.ClientNodes))
}

func ContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
	})
}

func (nd *NodeDiscovery) HttpRegisterNode(w http.ResponseWriter, r *http.Request) {
	fmt.Println("register")
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

	fmt.Fprint(w, []byte("{\"detail\": \"ok\"}"))
}

func HandleRequests(port int, nd *NodeDiscovery) {
	r := mux.NewRouter()
	r.Use(ContentTypeMiddleware)

	r.HandleFunc("/get-all", nd.HttpGetAllNodes).Methods("GET")
	r.HandleFunc("/get-blockchain", nd.HttpGetBlockchain).Methods("GET")
	r.HandleFunc("/get-clients", nd.HttpGetClients).Methods("GET")
	r.HandleFunc("/register", nd.HttpRegisterNode).Methods("POST")

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), r))
}

func main() {
	var nd NodeDiscovery
	portPtr := flag.Int("port", 9999, "HTTP listener port")

	fmt.Println("[Node Discovery] Starting HTTP Listener on port", *portPtr)
	HandleRequests(*portPtr, &nd)
}
