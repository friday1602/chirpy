package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// POST /api/refresh
// refreshTokenAuth authorizes user with refresh token on the database
// then sending new access-token to the user.
func (a *apiConfig) refreshTokenAuth(w http.ResponseWriter, r *http.Request) {

	token, err := validateToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if claims, ok := token.Claims.(*CustomClaims); ok {
		if !isRefreshToken(claims.Issuer) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		user, err := a.db.GetUserByID(claims.UserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		if user.RefreshToken != token.Raw {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}

		jwtSecret := os.Getenv("JWT_SECRET")
		claims.Issuer = "chirpy-access"
		claims.IssuedAt = jwt.NewNumericDate(time.Now())
		claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Hour))
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		stringToken, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			http.Error(w, "Error signstring token", http.StatusInternalServerError)
			return
		}
		resp, err := json.Marshal(struct {
			Token string `json:"token"`
		}{
			Token: stringToken,
		})
		if err != nil {
			http.Error(w, "Error marshalling json", http.StatusInternalServerError)
			return
		}
		w.Write(resp)

	}
}
