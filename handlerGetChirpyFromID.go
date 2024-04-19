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

	chirps, err := a.db.GetChirps()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// check if ID is in database ID range
	if ID > len(chirps) || ID <= 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	resp, err := json.Marshal(chirps[ID-1])
	if err != nil {
		http.Error(w, "Error marshalling json", http.StatusInternalServerError)
		return
	}
	w.Write(resp)

}
