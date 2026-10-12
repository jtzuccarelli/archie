package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jtzuccarelli/archie/internal/processor"
	"github.com/jtzuccarelli/archie/internal/store"
)

type Worker struct {
	store     *store.Store
	processor *processor.Processor
	logger    *slog.Logger
}

func New(logger *slog.Logger, store *store.Store, processor *processor.Processor) *Worker {
	return &Worker{
		store:     store,
		processor: processor,
		logger:    logger,
	}
}

func (w *Worker) Run(ctx context.Context) error {

	for {

		if ctx.Err() != nil {
			return ctx.Err()
		}

		found, err := w.processNext(ctx)

		if err != nil {
			return err
		}

		if found {
			continue
		}

		timer := time.NewTimer(5 * time.Second)

		select {

		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
			// wait here before querying again so we don't hammer db
		}
	}
}

func (w *Worker) processNext(ctx context.Context) (bool, error) {
	c, err := w.store.ClaimCalls(ctx, 1)

	if err != nil {
		return false, err
	}

	if len(c) == 0 {
		return false, nil
	}

	claimed := c[0]
	w.logger.Info("call claimed", "call_id", claimed.ID, "attempts", claimed.Attempts)

	outcome, err := w.processor.Process(ctx, claimed)
	if err != nil {
		return true, fmt.Errorf("processing call %d: %w", claimed.ID, err)
	}

	if err := w.store.CompleteCall(ctx, claimed.ID, claimed.Attempts, outcome); err != nil {
		return true, err
	}

	return true, nil
}
