package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Friday1602/chirpy/database"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

type apiConfig struct {
	fileserverHits int
}
type chripyParams struct {
	Body string `json:"body"`
}
type user struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds int    `json:"expires_in_seconds"`
}
type errorResponse struct {
	Error string `json:"error"`
}

type CustomClaims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	mux := http.NewServeMux()
	apiCfg := &apiConfig{}
	fileServer := http.FileServer(http.Dir("./app"))
	mux.Handle("/app/*", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", fileServer))) //* for wildcard

	mux.HandleFunc("GET /admin/metrics", apiCfg.metrics)

	mux.HandleFunc("/api/reset", apiCfg.reset)

	fileServer = http.FileServer(http.Dir("./app/assets"))
	mux.Handle("/app/assets/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/assets", fileServer)))

	mux.HandleFunc("GET /api/healthz", readiness)

	mux.HandleFunc("POST /api/chirps", validateChirpy)
	mux.HandleFunc("GET /api/chirps", getChirpy)
	mux.HandleFunc("GET /api/chirps/{chirpID}", getChirpyFromID)
	mux.HandleFunc("POST /api/users", createUser)
	mux.HandleFunc("POST /api/login", userValidation)
	mux.HandleFunc("PUT /api/users", updateUser)

	corsMux := middlewareCors(mux)
	log.Print("starting server on :8080")
	err = http.ListenAndServe(":8080", corsMux)
	log.Fatal(err)
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")

	jwtSecret := os.Getenv("JWT_SECRET")
	tokenFromHeader := authHeader[len("Bearer "):]
	token, err := jwt.ParseWithClaims(tokenFromHeader, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		http.Error(w, "Invalid Token", http.StatusUnauthorized)
		return
	}

	if claims, ok := token.Claims.(*CustomClaims); ok {
		db, err := database.NewUserDB("userDatabase.json")
		if err != nil {
			responseErrorInJsonBody(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		userReq := user{}
		err = json.NewDecoder(r.Body).Decode(&userReq)
		if err != nil {
			responseErrorInJsonBody(w, "Error decoding json", http.StatusBadRequest)
			return
		}

		cost := bcrypt.DefaultCost
		password, err := bcrypt.GenerateFromPassword([]byte(userReq.Password), cost)
		if err != nil {
			responseErrorInJsonBody(w, "Error updating password", http.StatusBadRequest)
			return
		}
		user, err := db.UpdateUserDB(claims.UserID, userReq.Email, password)
		if err != nil {
			responseErrorInJsonBody(w, "Error updating password", http.StatusInternalServerError)
			return
		}

		resp, err := json.Marshal(struct{
			Email string `json:"email"`
			ID int `json:"id"`
		}{
			Email: user.Email,
			ID: user.ID,
		})
		if err != nil {
			responseErrorInJsonBody(w, "Error marshalling json", http.StatusInternalServerError)
			return
		}
		w.Write(resp)
	} else {
		http.Error(w, "Unknow Claims type", http.StatusBadRequest)
		return
	}

}

// validate user logging in
func userValidation(w http.ResponseWriter, r *http.Request) {
	// decode request to struct
	userReq := user{}
	err := json.NewDecoder(r.Body).Decode(&userReq)
	if err != nil {
		responseErrorInJsonBody(w, "Something went wrong", http.StatusBadRequest)
		return
	}

	// connect to database
	db, err := database.NewUserDB("userDatabase.json")
	if err != nil {
		responseErrorInJsonBody(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// get users
	users, err := db.GetUser()
	if err != nil {
		responseErrorInJsonBody(w, "Internal Server Error", http.StatusInternalServerError)
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

			timeToExpire := time.Hour * 24
			if userReq.ExpiresInSeconds != 0 {
				timeToExpire = time.Second * time.Duration(userReq.ExpiresInSeconds)
			}
			// create claims
			claims := CustomClaims{
				UserID: user.ID,
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:    "chirpy",
					IssuedAt:  jwt.NewNumericDate(time.Now()),
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(timeToExpire)),
					Subject:   strconv.Itoa(user.ID),
				},
			}
			// get jwt secret from .env file
			jwtSecret := os.Getenv("JWT_SECRET")
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			ss, err := token.SignedString([]byte(jwtSecret))
			if err != nil {
				fmt.Println(jwtSecret, err)
				responseErrorInJsonBody(w, "Error creating token", http.StatusInternalServerError)
				return
			}

			resp, err := json.Marshal(struct {
				ID    int    `json:"id"`
				Email string `json:"email"`
				Token string `json:"token"`
			}{
				ID:    user.ID,
				Email: user.Email,
				Token: ss,
			})
			if err != nil {
				responseErrorInJsonBody(w, "Error mashalling json", http.StatusInternalServerError)
				return
			}
			w.Write(resp)
		}
	}
}

// create users
func createUser(w http.ResponseWriter, r *http.Request) {
	//decode request json to user struct
	userReq := user{}
	err := json.NewDecoder(r.Body).Decode(&userReq)
	if err != nil {
		responseErrorInJsonBody(w, "Something went wrong", http.StatusBadRequest)
		return
	}

	// hash the password using bcrypt
	cost := bcrypt.DefaultCost
	password, err := bcrypt.GenerateFromPassword([]byte(userReq.Password), cost)
	if err != nil {
		responseErrorInJsonBody(w, "Error creating user", http.StatusBadRequest)
		return
	}

	// initiated user database
	db, err := database.NewUserDB("userDatabase.json")
	if err != nil {
		responseErrorInJsonBody(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// check if user exists
	users, err := db.GetUser()
	if err != nil {
		responseErrorInJsonBody(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	for _, user := range users {
		if user.Email == userReq.Email {
			responseErrorInJsonBody(w, "This Email already exists", http.StatusBadRequest)
			return
		}
	}

	// create new user
	createdDB, err := db.CreateUser(userReq.Email, password)
	if err != nil {
		responseErrorInJsonBody(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// create completed response with 201 and encoding user data from database
	// using anonymous struct to response specific field (exclude password)
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(struct {
		Email string `json:"email"`
		ID    int    `json:"id"`
	}{
		Email: createdDB.Email,
		ID:    createdDB.ID,
	})
	if err != nil {
		responseErrorInJsonBody(w, "Error Encoding json", http.StatusInternalServerError)
		return
	}

}

// validate if chirpy is valid. if valid response json valid body. if not response json error body
func validateChirpy(w http.ResponseWriter, r *http.Request) {

	// decode json body and check for error
	chirpyParam := chripyParams{}
	err := json.NewDecoder(r.Body).Decode(&chirpyParam)
	if err != nil {
		responseErrorInJsonBody(w, "Something went wrong", http.StatusBadRequest)
		return
	}

	// check if json body length is more than 140 characters long.
	if len([]rune(chirpyParam.Body)) > 140 {
		responseErrorInJsonBody(w, "Chirp is too long", http.StatusBadRequest)
		return
	}
	// replace all profanes with ****
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	stringChirpy := strings.Split(chirpyParam.Body, " ")
	for i, word := range stringChirpy {
		for _, badWord := range badWords {
			if strings.ToLower(word) == badWord {
				stringChirpy[i] = "****"
			}
		}
	}
	cleanedChirpy := strings.Join(stringChirpy, " ")
	db, err := database.NewDB("database.json")
	if err != nil {
		responseErrorInJsonBody(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	createdDB, err := db.CreateChirp(cleanedChirpy)
	if err != nil {
		responseErrorInJsonBody(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// chirp is valid response valid successReponse struct encoded to json
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(&createdDB)
	if err != nil {
		responseErrorInJsonBody(w, "Error Encoding json", http.StatusInternalServerError)
		return
	}

}

func getChirpy(w http.ResponseWriter, r *http.Request) {
	db, err := database.NewDB("database.json")
	if err != nil {
		responseErrorInJsonBody(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	chirps, err := db.GetChirps()
	if err != nil {
		responseErrorInJsonBody(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	resp, err := json.Marshal(chirps)
	if err != nil {
		responseErrorInJsonBody(w, "Error marshalling json", http.StatusInternalServerError)
		return
	}
	w.Write(resp)
}

// get chirpy from specific ID
func getChirpyFromID(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	ID, err := strconv.Atoi(chirpID)
	if err != nil {
		responseErrorInJsonBody(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	db, err := database.NewDB("database.json")
	if err != nil {
		responseErrorInJsonBody(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	chirps, err := db.GetChirps()
	if err != nil {
		responseErrorInJsonBody(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// check if ID is in database ID range
	if ID > len(chirps) || ID <= 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	resp, err := json.Marshal(chirps[ID-1])
	if err != nil {
		responseErrorInJsonBody(w, "Error marshalling json", http.StatusInternalServerError)
		return
	}
	w.Write(resp)

}

// response specific error message encode to json body if any error occurs.
func responseErrorInJsonBody(w http.ResponseWriter, errorMessage string, statusCode int) {
	errorResp, err := json.Marshal(errorResponse{Error: errorMessage})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error marshaling json: %s", err)
		return
	}
	w.WriteHeader(statusCode)
	w.Write(errorResp)
}

// middlewareMetrics gathers amout of request to the page
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})

}

// metrics prints counts to the body
func (cfg *apiConfig) metrics(w http.ResponseWriter, r *http.Request) {
	hits := cfg.fileserverHits
	tmpl := `
	<!DOCTYPE html>
	<html>
	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited {{.}} times!</p>
	</body>
	
	</html>
	`
	t, err := template.New("admin").Parse(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if err := t.Execute(w, hits); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// reset resets counts
func (cfg *apiConfig) reset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits = 0
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func readiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}
