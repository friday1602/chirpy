package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// POST /api/refresh
func (a *apiConfig) refreshTokenAuth(w http.ResponseWriter, r *http.Request) {
	authString := r.Header.Get("Authorization")

	jwtSecret := os.Getenv("JWT_SECRET")
	tokenFromHeader := authString[len("Bearer "):]
	token, err := jwt.ParseWithClaims(tokenFromHeader, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		http.Error(w, "Invalid Token", http.StatusUnauthorized)
		return
	}
	if claims, ok := token.Claims.(*CustomClaims); ok {
		if claims.Issuer != "chirpy-refresh" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		users, err := a.db.GetUser()
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		if users[claims.UserID - 1].RefreshToken != tokenFromHeader {
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}

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
