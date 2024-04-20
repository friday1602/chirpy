package main

import (
	"net/http"
)

// POST /api/revoke
func (a *apiConfig) revokeToken(w http.ResponseWriter, r *http.Request) {

	token, err := validateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}

	if claims, ok := token.Claims.(*CustomClaims); ok {
		if claims.Issuer != "chirpy-refresh" {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		// revoke the refresh token in the database
		err := a.db.RevokeToken(claims.UserID)
		if err != nil {
			http.Error(w, "Error revoking token", http.StatusInternalServerError)
			return
		}
	}
}
