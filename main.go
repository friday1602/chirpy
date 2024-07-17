package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/friday1602/chirpy/database"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

type apiConfig struct {
	fileserverHits int
	db             *database.DB
	chirpyDatabase *database.DB
	infoLog        *log.Logger
	errorLog       *log.Logger
}
type chripyParams struct {
	Body string `json:"body"`
}
type user struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CustomClaims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

func main() {
	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime)

	if *dbg {
		err := os.Remove("chirpyDatabase.json")
		if err != nil {
			errorLog.Fatal(err)
		}

		err = os.Remove("userDatabase.json")
		if err != nil {
			errorLog.Fatal(err)
		}
	}

	err := godotenv.Load()
	if err != nil {
		errorLog.Fatal("error loading .env file")
	}

	apiCfg := &apiConfig{
		errorLog: errorLog,
		infoLog:  infoLog,
	}

	apiCfg.db, err = database.NewUserDB("userDatabase.json")
	if err != nil {
		errorLog.Fatal(err)
	}
	apiCfg.chirpyDatabase, err = database.NewDB("chirpyDatabase.json")
	if err != nil {
		errorLog.Fatal(err)
	}

	port := os.Getenv("PORT")
	infoLog.Print("starting server on :", port)
	srv := http.Server{
		Addr:              ":" + port,
		Handler:           apiCfg.routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ErrorLog:          errorLog,
	}
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}
