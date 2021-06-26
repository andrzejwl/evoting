package pow

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func HandleRequests(port int, blockchain *Blockchain) {
	r := mux.NewRouter()

	r.HandleFunc("/chain/", blockchain.HttpGetChain).Methods("GET")
	r.HandleFunc("/transaction/create", blockchain.HttpCreateTransaction).Methods("POST")
	r.HandleFunc("/update", blockchain.HttpUpdate).Methods("POST")
	r.HandleFunc("/debug/update", blockchain.HttpTriggerUpdate).Methods("GET")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), r))
}
