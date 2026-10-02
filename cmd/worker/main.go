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
	"github.com/jtzuccarelli/archie/internal/analyzer"
	"github.com/jtzuccarelli/archie/internal/processor"
	"github.com/jtzuccarelli/archie/internal/store"
	"github.com/jtzuccarelli/archie/internal/trackdrive"
	"github.com/jtzuccarelli/archie/internal/transcriber"
	"github.com/jtzuccarelli/archie/internal/worker"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := run(logger); err != nil {
		logger.Error("worker exited", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {

	if err := godotenv.Load(); err != nil {
		logger.Error("no .env file loaded for worker", "error", err)
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

	st := store.NewStore(pool)
	c := http.Client{Timeout: 60 * time.Second}
	td := trackdrive.New(cfg.tdAuthHeader, &c)
	tr := transcriber.New(cfg.openaiApiKey, &c)
	an := analyzer.New(cfg.openaiApiKey, &c)
	proc := processor.New(td, tr, an)

	w := worker.New(logger, st, proc)

	return w.Run(ctx)

}
