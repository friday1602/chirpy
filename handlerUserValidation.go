package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// validate user logging in POST /api/login
func (a *apiConfig) userValidation(w http.ResponseWriter, r *http.Request) {
	// decode request to struct
	userReq := user{}
	err := json.NewDecoder(r.Body).Decode(&userReq)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusBadRequest)
		return
	}

	// get users
	users, err := a.db.GetUser()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// range all users in database compare to the request
	for _, user := range users {
		if user.Email == userReq.Email {
			err := bcrypt.CompareHashAndPassword(user.Password, []byte(userReq.Password))
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			timeToExpireAccessToken := time.Hour            // 1 Hour
			timeToExpireRefreshToken := time.Hour * 24 * 60 // 60 Days

			// create claims for access token
			claimsAccessToken := CustomClaims{
				UserID: user.ID,
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:    "chirpy-access",
					IssuedAt:  jwt.NewNumericDate(time.Now()),
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(timeToExpireAccessToken)),
					Subject:   strconv.Itoa(user.ID),
				},
			}

			// create claims for refresh token
			claimsRefreshToken := CustomClaims{
				UserID: user.ID,
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:    "chirpy-refresh",
					IssuedAt:  jwt.NewNumericDate(time.Now()),
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(timeToExpireRefreshToken)),
					Subject:   strconv.Itoa(user.ID),
				},
			}
			// get jwt secret from .env file
			jwtSecret := os.Getenv("JWT_SECRET")
			// create access and refresh tokens
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsAccessToken)
			refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsRefreshToken)

			signedStringToken, err := token.SignedString([]byte(jwtSecret))
			if err != nil {
				fmt.Println(jwtSecret, err)
				http.Error(w, "Error creating token", http.StatusInternalServerError)
				return
			}

			signedStringRefreshToken, err := refreshToken.SignedString([]byte(jwtSecret))
			if err != nil {
				fmt.Println(jwtSecret, err)
				http.Error(w, "Error creating token", http.StatusInternalServerError)
				return
			}

			err = a.db.StoreToken(user.ID, signedStringRefreshToken)
			if err != nil {
				http.Error(w, "error storing refresh token", http.StatusInternalServerError)
				return
			}

			resp, err := json.Marshal(struct {
				Token        string `json:"token"`
				RefreshToken string `json:"refresh_token"`
			}{
				Token:        signedStringToken,
				RefreshToken: signedStringRefreshToken,
			})
			if err != nil {
				http.Error(w, "Error mashalling json", http.StatusInternalServerError)
				return
			}
			w.Write(resp)
			return
		}
	}
	http.Error(w, "User not found", http.StatusNotFound)
}
