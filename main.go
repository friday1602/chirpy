package main

import (
	"log"
	"net/http"
)

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func readiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func main() {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./app"))
	mux.Handle("/app/", http.StripPrefix("/app", fileServer))

	fileServer = http.FileServer(http.Dir("./app/assets"))
	mux.Handle("/app/assets/", http.StripPrefix("/app/assets", fileServer))

	mux.HandleFunc("/healthz", readiness)

	corsMux := middlewareCors(mux)

	log.Print("starting server on :8080")
	err := http.ListenAndServe(":8080", corsMux)
	log.Fatal(err)
}
