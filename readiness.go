package main

import (
	"log"
	"net/http"
)

func readiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "")
	w.WriteHeader(200)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		log.Print(err)
	}
}
