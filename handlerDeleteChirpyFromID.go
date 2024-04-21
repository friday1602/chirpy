package main

import "net/http"

func(a *apiConfig) deleteChirpyFromID(w http.ResponseWriter, r *http.Request) {
	token, err := validateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
	}

	if claims, ok := token.Claims.(*CustomClaims); ok {
		userID := claims.UserID
		
	}
}