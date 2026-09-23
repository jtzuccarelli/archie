package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	slog.Info("booting archie api...")

	if err := godotenv.Load(); err != nil {
		slog.Debug("no .env file loaded", "error", err)
	}

	mux := http.NewServeMux()
	port := os.Getenv("PORT")

	if port == "" {
		port = "4000"
	}

	slog.Info("starting server on", "port", port)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		slog.Error("Error starting server...", "error", err)
		os.Exit(1)
	}

}
