package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (a *apiConfig) getChirpy(w http.ResponseWriter, r *http.Request) {
	autherID := r.URL.Query().Get("auther_id")

	chirps, err := a.chirpyDatabase.GetChirps()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if autherID != "" {
		authID, err := strconv.Atoi(autherID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		chirps, err = a.chirpyDatabase.GetChirpsByAuthorID(authID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		resp, err := json.Marshal(chirps)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(resp)
		return
	}

	// if auther id is not provided ... 
	resp, err := json.Marshal(chirps)
	if err != nil {
		http.Error(w, "Error marshalling json", http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}
