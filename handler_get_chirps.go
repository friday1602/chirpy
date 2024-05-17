package main

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/friday1602/chirpy/database"
)

func (a *apiConfig) getChirpy(w http.ResponseWriter, r *http.Request) {
	authorID := r.URL.Query().Get("author_id")
	sortChirp := r.URL.Query().Get("sort")

	var chirps []database.Chirp
	var err error
	if authorID != "" {
		authID, err := strconv.Atoi(authorID)
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

	for _, c := range chirps {
		fmt.Fprintf(w, "%s\n", c.Body)
	}

}
