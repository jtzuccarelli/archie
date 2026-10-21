package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/jtzuccarelli/archie/internal/store"
)

type Sweeper struct {
	store    *store.Store
	logger   *slog.Logger
	interval time.Duration
}

func NewSweeper(logger *slog.Logger, store *store.Store, interval time.Duration) *Sweeper {
	return &Sweeper{
		store:    store,
		logger:   logger,
		interval: interval,
	}
}

func (s *Sweeper) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		s.sweep(ctx)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (s *Sweeper) sweep(ctx context.Context) {
	retried, failed, err := s.store.RecoverStaleCalls(ctx)
	if err != nil {
		s.logger.Error("sweep failed", "error", err)
		return
	}

	if retried > 0 || failed > 0 {
		s.logger.Info("recovered stale calls", "retried", retried, "failed", failed)
	}
}
