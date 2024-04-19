package main

import (
	"encoding/json"
	"net/http"
)

func (a *apiConfig) getChirpy(w http.ResponseWriter, r *http.Request) {

	chirps, err := a.db.GetChirps()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	resp, err := json.Marshal(chirps)
	if err != nil {
		http.Error(w, "Error marshalling json", http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}
