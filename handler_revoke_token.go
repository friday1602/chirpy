package main

import (
	"net/http"
)

// POST /api/revoke
// revokeToken revokes the refresh-token in the database
// then creates and sends new refresh-token to the user and the database.
func (a *apiConfig) revokeToken(w http.ResponseWriter, r *http.Request) {

	token, err := validateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}

	if claims, ok := token.Claims.(*CustomClaims); ok {
		if !isRefreshToken(claims.Issuer) {
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
