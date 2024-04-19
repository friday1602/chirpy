package main

import (
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

// POST /api/revoke
func (a *apiConfig) revokeToken(w http.ResponseWriter, r *http.Request) {
	authString := r.Header.Get("Authorization")
	tokenFromHeader := authString[len("Bearer "):]

	jwtSecret := os.Getenv("JWT_SECRET")
	jwtToken, err := jwt.ParseWithClaims(tokenFromHeader, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}
	if claims, ok := jwtToken.Claims.(*CustomClaims); ok {
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
