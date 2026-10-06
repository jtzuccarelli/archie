package main

import (
	"crypto/subtle"
	"net/http"
)

const qvSecretHeader = "X-QV-Webhook-Secret"

func (app *application) requireQVSecret(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get(qvSecretHeader)
		if subtle.ConstantTimeCompare([]byte(got), app.qvSecret) != 1 {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
