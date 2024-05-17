package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

type webhooksRequest struct {
	Event string `json:"event"`
	Data  struct {
		UserID int `json:"user_id"`
	} `json:"data"`
}

func (a *apiConfig) upgradeToRedChirpy(w http.ResponseWriter, r *http.Request) {
	apiAuth := r.Header.Get("Authorization")
	apiKeys := strings.Split(apiAuth, " ")

	if len(apiKeys) != 2 || apiKeys[0] != "ApiKey" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	polkaKey := os.Getenv("POLKA_API_KEY")
	if polkaKey != apiKeys[1] {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	webhooksReq := webhooksRequest{}
	err := json.NewDecoder(r.Body).Decode(&webhooksReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if webhooksReq.Event != "user.upgraded" {
		return
	}

	err = a.db.UpgradeUser(webhooksReq.Data.UserID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	}

}
