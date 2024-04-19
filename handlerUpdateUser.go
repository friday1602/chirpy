package main

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// PUT /api/users endpoint
func (a *apiConfig) updateUser(w http.ResponseWriter, r *http.Request) {
	// get token from auth header
	authHeader := r.Header.Get("Authorization")

	jwtSecret := os.Getenv("JWT_SECRET")
	tokenFromHeader := authHeader[len("Bearer "):] // trim Bearer from auth.. the rest is token
	token, err := jwt.ParseWithClaims(tokenFromHeader, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		http.Error(w, "Invalid Token", http.StatusUnauthorized)
		return
	}

	if claims, ok := token.Claims.(*CustomClaims); ok {
		if claims.Issuer != "chirpy-access" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userReq := user{}
		err = json.NewDecoder(r.Body).Decode(&userReq)
		if err != nil {
			http.Error(w, "Error decoding json", http.StatusBadRequest)
			return
		}

		cost := bcrypt.DefaultCost
		password, err := bcrypt.GenerateFromPassword([]byte(userReq.Password), cost)
		if err != nil {
			http.Error(w, "Error updating password", http.StatusBadRequest)
			return
		}
		user, err := a.db.UpdateUserDB(claims.UserID, userReq.Email, password)
		if err != nil {
			http.Error(w, "Error updating password", http.StatusInternalServerError)
			return
		}

		resp, err := json.Marshal(struct {
			Email string `json:"email"`
			ID    int    `json:"id"`
		}{
			Email: user.Email,
			ID:    user.ID,
		})
		if err != nil {
			http.Error(w, "Error marshalling json", http.StatusInternalServerError)
			return
		}
		w.Write(resp)
	} else {
		http.Error(w, "Unknow Claims type", http.StatusBadRequest)
		return
	}

}
