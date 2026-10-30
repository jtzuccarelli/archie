package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jtzuccarelli/archie/internal/user"
)

type loginRequest struct {
	Id       int64  `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (app *application) loginHandler(w http.ResponseWriter, r *http.Request) {
	app.logger.InfoContext(r.Context(), "login endpoint hit")

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password are required", http.StatusBadRequest)
		return
	}

	u, err := app.store.GetUserByEmail(r.Context(), req.Email)

	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err != nil {
		app.logger.ErrorContext(r.Context(), "looking up user", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := user.VerifyPassword(u.PasswordHash, req.Password); err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := user.GenerateSessionToken()
	if err != nil {
		app.logger.ErrorContext(r.Context(), "generating session token", "user_id", u.ID, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	tokenHash := user.HashSessionToken(token)
	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	if err := app.store.SaveUserSession(r.Context(), u.ID, tokenHash, expiresAt); err != nil {
		app.logger.ErrorContext(r.Context(), "saving user session", "user_id", u.ID, "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
