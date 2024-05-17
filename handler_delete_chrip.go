package main

import (
	"net/http"
	"strconv"
)

// DELETE /api/chirps/{chirpID}
// deleteChirpyFromID delete chirpy from specific id
// authoriztion before deletion
func(a *apiConfig) deleteChirpyFromID(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	ID, err := strconv.Atoi(chirpID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := validateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if claims, ok := token.Claims.(*CustomClaims); ok {
		if !isAcessToken(claims.Issuer) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		userID := claims.UserID
		err := a.chirpyDatabase.DeleteDB(userID, ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
	}
}