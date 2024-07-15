package main

import "net/http"

func (apiCfg *apiConfig) routes() *http.ServeMux {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("./app"))
	mux.Handle("/app/*", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", fileServer))) //* for wildcard
	mux.HandleFunc("GET /admin/metrics", apiCfg.metrics)

	mux.HandleFunc("/api/reset", apiCfg.reset)

	fileServer = http.FileServer(http.Dir("./app/assets"))
	mux.Handle("/app/assets/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/assets", fileServer)))

	mux.HandleFunc("GET /api/healthz", readiness)

	mux.HandleFunc("POST /api/chirps", apiCfg.validateChirpy)
	mux.HandleFunc("GET /api/chirps", apiCfg.getChirpy)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.getChirpyFromID)
	mux.HandleFunc("POST /api/users", apiCfg.createUser)
	mux.HandleFunc("POST /api/login", apiCfg.userValidation)
	mux.HandleFunc("PUT /api/users", apiCfg.updateUser)
	mux.HandleFunc("POST /api/refresh", apiCfg.refreshTokenAuth)
	mux.HandleFunc("POST /api/revoke", apiCfg.revokeToken)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.deleteChirpyFromID)
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.upgradeToRedChirpy)

	return mux
}
