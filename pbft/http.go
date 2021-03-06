package pbft

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func HandleRequests(port int, blockchain *Blockchain) {
	r := mux.NewRouter()

	r.HandleFunc("/pre-prepare", blockchain.HttpPrePrepare).Methods("POST")
	r.HandleFunc("/prepare", blockchain.HttpPrepare).Methods("POST")
	r.HandleFunc("/chain", blockchain.HttpGetChain).Methods("GET")
	r.HandleFunc("/request", blockchain.HttpRequest).Methods("POST")
	r.HandleFunc("/pending", blockchain.HttpGetPending).Methods("GET")
	r.HandleFunc("/peers", blockchain.HttpGetPeers).Methods("GET")
	r.HandleFunc("/refresh", blockchain.HttpTriggerRefresh).Methods("GET")

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), r))
}
