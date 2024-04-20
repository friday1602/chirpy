package main

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// validateToken checks validity of token from header.
// it returns string token if it is valid or error if it is not.
func validateToken(r *http.Request) (*jwt.Token, error) {
	authHeader := r.Header.Get("Authorization")

	jwtSecret := os.Getenv("JWT_SECRET")

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, errors.New("invalid token")
	}
	tokenFromHeader := parts[1]

	token, err := jwt.ParseWithClaims(tokenFromHeader, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}
