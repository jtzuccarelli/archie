package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/jtzuccarelli/archie/internal/call"
)

type application struct {
	logger *slog.Logger
	store  *call.Store
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := run(logger); err != nil {
		logger.Error("api exited", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	if err := godotenv.Load(); err != nil {
		logger.Debug("no .env file loaded", "error", err)
	}

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.dsn)
	if err != nil {
		return fmt.Errorf("creating connection pool: %w", err)
	}

	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("pinging database: %w", err)
	}

	app := &application{
		logger: logger,
		store:  call.NewStore(pool),
	}

	srv := &http.Server{
		Addr:         ":" + cfg.port,
		Handler:      app.routes(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	app.logger.Info("starting server", "port", cfg.port)

	return srv.ListenAndServe()
}
