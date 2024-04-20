package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

// validate if chirpy is valid. if valid response json valid body. if not response json error body
// POST /api/chrips
func (a *apiConfig) validateChirpy(w http.ResponseWriter, r *http.Request) {

	// decode json body and check for error
	chirpyParam := chripyParams{}
	err := json.NewDecoder(r.Body).Decode(&chirpyParam)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusBadRequest)
		return
	}

	// check if json body length is more than 140 characters long.
	if len([]rune(chirpyParam.Body)) > 140 {
		http.Error(w, "Chirp is too long", http.StatusBadRequest)
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
	createdDB, err := a.db.CreateChirp(cleanedChirpy)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// chirp is valid response valid successReponse struct encoded to json
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(&createdDB)
	if err != nil {
		http.Error(w, "Error Encoding json", http.StatusInternalServerError)
		return
	}

}
