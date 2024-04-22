package main

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// create users POST /api/users
func (a *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	//decode request json to user struct
	userReq := user{}
	err := json.NewDecoder(r.Body).Decode(&userReq)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusBadRequest)
		return
	}

	// hash the password using bcrypt
	cost := bcrypt.DefaultCost
	password, err := bcrypt.GenerateFromPassword([]byte(userReq.Password), cost)
	if err != nil {
		http.Error(w, "Error creating user", http.StatusBadRequest)
		return
	}

	// check if user exists
	users, err := a.db.GetUser()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	for _, user := range users {
		if user.Email == userReq.Email {
			http.Error(w, "This Email already exists", http.StatusBadRequest)
			return
		}
	}

	// create new user
	createdDB, err := a.db.CreateUser(userReq.Email, password)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// create completed response with 201 and encoding user data from database
	// using anonymous struct to response specific field (exclude password)
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(struct {
		Email string `json:"email"`
		ID    int    `json:"id"`
		IsChirpyRed bool `json:"is_chirpy_red"`
	}{
		Email: createdDB.Email,
		ID:    createdDB.ID,
		IsChirpyRed: createdDB.IsChirpyRed,
	})
	if err != nil {
		http.Error(w, "Error Encoding json", http.StatusInternalServerError)
		return
	}

}
