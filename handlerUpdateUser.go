package main

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// PUT /api/users endpoint
func (a *apiConfig) updateUser(w http.ResponseWriter, r *http.Request) {
	// get token from auth header
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
