package main

import (
	"encoding/json"
	"net/http"
	"sort"
	"strconv"

	"github.com/friday1602/chirpy/database"
)

func (a *apiConfig) getChirpy(w http.ResponseWriter, r *http.Request) {
	autherID := r.URL.Query().Get("auther_id")
	sortChirp := r.URL.Query().Get("sort")

	var chirps []database.Chirp
	var err error
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

	} else {
		chirps, err = a.chirpyDatabase.GetChirps()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

	}

	if sortChirp == "desc" {
		sort.Slice(chirps, func(i, j int) bool { return chirps[i].ID > chirps[j].ID })
	}

	resp, err := json.Marshal(chirps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(resp)

}
