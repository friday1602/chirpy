package main

import "net/http"

func readiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}
