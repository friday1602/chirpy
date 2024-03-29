package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
)

type apiConfig struct {
	fileserverHits int
}
type chripyParams struct {
	Body string `json:"body"`
}
type errorResponse struct {
	Error string `json:"error"`
}
type successReponse struct {
	Valid bool `json:"valid"`
}

func main() {
	mux := http.NewServeMux()
	apiCfg := &apiConfig{}
	fileServer := http.FileServer(http.Dir("./app"))
	mux.Handle("/app/*", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", fileServer))) //* for wildcard

	mux.HandleFunc("GET /admin/metrics", apiCfg.metrics)

	mux.HandleFunc("/api/reset", apiCfg.reset)

	fileServer = http.FileServer(http.Dir("./app/assets"))
	mux.Handle("/app/assets/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/assets", fileServer)))

	mux.HandleFunc("GET /api/healthz", readiness)

	mux.HandleFunc("POST /api/validate_chirp", validateChirpy)

	corsMux := middlewareCors(mux)

	log.Print("starting server on :8080")
	err := http.ListenAndServe(":8080", corsMux)
	log.Fatal(err)
}

// validate if chirpy is valid. if valid response json valid body. if not response json error body
func validateChirpy(w http.ResponseWriter, r *http.Request) {

	// decode json body and check for error
	chirpyParam := chripyParams{}
	err := json.NewDecoder(r.Body).Decode(&chirpyParam)
	if err != nil {
		responseErrorInJsonBody(w, "Something went wrong", http.StatusBadRequest)
		return
	}

	// check if json body length is more than 140 characters long.
	if len([]rune(chirpyParam.Body)) > 140 {
		responseErrorInJsonBody(w, "Chirp is too long", http.StatusBadRequest)
		return
	}

	// chirp is valid response valid successReponse struct encoded to json
	json.NewEncoder(w).Encode(successReponse{Valid: true})
}

// response specific error message encode to json body if any error occurs.
func responseErrorInJsonBody(w http.ResponseWriter, errorMessage string, statusCode int) {
	errorResp, err := json.Marshal(errorResponse{Error: errorMessage})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error marshaling json: %s", err)
		return
	}
	w.WriteHeader(statusCode)
	w.Write(errorResp)
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
