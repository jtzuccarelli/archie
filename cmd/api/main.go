package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type application struct {
	logger *slog.Logger
}

func main() {

	app := &application{
		logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}

	if err := godotenv.Load(); err != nil {
		app.logger.Debug("no .env file loaded", "error", err)
	}

	port := os.Getenv("PORT")

	if port == "" {
		port = "4000"
	}

	app.logger.Info("starting server on", "port", port)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      app.routes(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		app.logger.Error("Error starting server...", "error", err)
		os.Exit(1)
	}

}
