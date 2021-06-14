package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func handleRequests(port int, blockchain *Blockchain) {
	r := mux.NewRouter()

	r.HandleFunc("/chain/", blockchain.HttpGetChain).Methods("GET")
	r.HandleFunc("/transaction/create", blockchain.HttpCreateTransaction).Methods("POST")
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), r))
}
