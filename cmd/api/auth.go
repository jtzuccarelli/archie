package main

import (
	"crypto/subtle"
	"net/http"

	"github.com/jtzuccarelli/archie/internal/user"
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

func (app *application) requireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil || cookie.Value == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		tokenHash := user.HashSessionToken(cookie.Value)
		valid, err := app.store.HasValidSession(r.Context(), tokenHash)
		if err != nil {
			app.logger.ErrorContext(r.Context(), "checking user session", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if !valid {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
