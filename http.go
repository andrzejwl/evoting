package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func handleRequests(blockchain *Blockchain) {
	r := mux.NewRouter()

	r.HandleFunc("/chain/", blockchain.HttpGetChain).Methods("GET")
	r.HandleFunc("/transaction/create", blockchain.HttpCreateTransaction).Methods("POST")
	log.Fatal(http.ListenAndServe(":5000", r))
}
