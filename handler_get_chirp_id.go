package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

// get chirpy from specific ID
func (a *apiConfig) getChirpyFromID(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	ID, err := strconv.Atoi(chirpID)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	chirp, err := a.chirpyDatabase.GetChirpyFromID(ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	resp, err := json.Marshal(chirp)
	if err != nil {
		http.Error(w, "Error marshalling json", http.StatusInternalServerError)
		return
	}
	w.Write(resp)

}
