package main

import (
	"html/template"
	"net/http"
)

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
