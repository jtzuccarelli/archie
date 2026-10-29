package main

import (
	"crypto/subtle"
	"net/http"
)

const qvSecretHeader = "X-QV-Webhook-Secret"

func (app *application) requireAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get(qvSecretHeader)
		if app.apiKey == "" || subtle.ConstantTimeCompare([]byte(got), []byte(app.apiKey)) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
