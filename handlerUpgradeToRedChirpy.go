package main

import (
	"encoding/json"
	"net/http"
)

type webhooksRequest struct {
	Event string `json:"event"`
	Data  struct {
		UserID int `json:"user_id"`
	} `json:"data"`
}

func (a *apiConfig) upgradeToRedChirpy(w http.ResponseWriter, r *http.Request) {
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
