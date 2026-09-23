package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	slog.Info("booting archie api...")

	mux := http.NewServeMux()

	slog.Info("starting server on port :4000")

	srv := &http.Server{
		Addr:         ":4000",
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
