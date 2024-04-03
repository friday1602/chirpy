package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Friday1602/chirpy/database"
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

	mux.HandleFunc("POST /api/chirps", validateChirpy)
	mux.HandleFunc("GET /api/chirps", getChirpy)
	mux.HandleFunc("GET /api/chirps/{chirpID}", getChirpyFromID)

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
	// replace all profanes with ****
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	stringChirpy := strings.Split(chirpyParam.Body, " ")
	for i, word := range stringChirpy {
		for _, badWord := range badWords {
			if strings.ToLower(word) == badWord {
				stringChirpy[i] = "****"
			}
		}
	}
	cleanedChirpy := strings.Join(stringChirpy, " ")
	db, err := database.NewDB("database.json")
	if err != nil {
		responseErrorInJsonBody(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	createdDB, err := db.CreateChirp(cleanedChirpy)
	if err != nil {
		responseErrorInJsonBody(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// chirp is valid response valid successReponse struct encoded to json
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(&createdDB)

}

func getChirpy(w http.ResponseWriter, r *http.Request) {
	db, err := database.NewDB("database.json")
	if err != nil {
		responseErrorInJsonBody(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	chirps, err := db.GetChirps()
	if err != nil {
		responseErrorInJsonBody(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	resp, err := json.Marshal(chirps)
	if err != nil {
		responseErrorInJsonBody(w, "Error marshalling json", http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}

// get chirpy from specific ID
func getChirpyFromID(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	ID, err := strconv.Atoi(chirpID)
	if err != nil {
		responseErrorInJsonBody(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	db, err := database.NewDB("database.json")
	if err != nil {
		responseErrorInJsonBody(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	chirps, err := db.GetChirps()
	if err != nil {
		responseErrorInJsonBody(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// check if ID is in database ID range
	if ID > len(chirps) || ID <= 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	resp, err := json.Marshal(chirps[ID-1])
	if err != nil {
		responseErrorInJsonBody(w, "Error marshalling json", http.StatusInternalServerError)
		return
	}
	w.Write(resp)

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
