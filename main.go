package main

import (
	"html/template"
	"log"
	"net/http"
)

type apiConfig struct {
	fileserverHits int
}

func main() {
	mux := http.NewServeMux()
	apiCfg := &apiConfig{}
	fileServer := http.FileServer(http.Dir("./app"))
	mux.Handle("/app/*", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", fileServer))) //* for wildcard

	mux.HandleFunc("GET /admin/metrics", apiCfg.metrics)

	mux.HandleFunc("/api/reset", apiCfg.reset)

	fileServer = http.FileServer(http.Dir("./app/assets"))
	mux.Handle("/app/assets/", http.StripPrefix("/app/assets", fileServer))

	mux.HandleFunc("GET /api/healthz", readiness)

	corsMux := middlewareCors(mux)

	log.Print("starting server on :8080")
	err := http.ListenAndServe(":8080", corsMux)
	log.Fatal(err)
}

// middlewareMetrics gathers amout of request to the page
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})

}

// metrics prints counts to the body
func (cfg *apiConfig) metrics(w http.ResponseWriter, r *http.Request) {
	hits := cfg.fileserverHits
	tmpl := `
	<!DOCTYPE html>
	<html>
	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited {{.}} times!</p>
	</body>
	
	</html>
	`
	t, err := template.New("admin").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if err := t.Execute(w, hits); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// reset resets counts
func (cfg *apiConfig) reset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = 0
}

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
	w.Header().Set("Content-Type", "")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}
