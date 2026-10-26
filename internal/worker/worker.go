package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jtzuccarelli/archie/internal/processor"
	"github.com/jtzuccarelli/archie/internal/retry"
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
	w.logger.InfoContext(ctx, "call claimed", "call_id", claimed.ID, "attempts", claimed.Attempts)

	outcome, err := w.processor.Process(ctx, claimed)
	maxAttemptsReached := claimed.Attempts >= 3

	if ctx.Err() != nil {
		return true, ctx.Err()
	}

	if err != nil {

		if !retry.IsRetryable(err) || maxAttemptsReached {
			e := w.store.FailCall(ctx, claimed.ID, claimed.Attempts, err)

			if errors.Is(e, store.ErrStaleClaim) {
				return true, nil
			}

			if e != nil {
				return true, fmt.Errorf("processing call %d: %w", claimed.ID, e)
			}

			w.logger.ErrorContext(ctx, "call failed",
				"call_id", claimed.ID,
				"attempts", claimed.Attempts,
				"error", err,
			)

			return true, nil
		} else {
			e := w.store.RetryCall(ctx, claimed.ID, claimed.Attempts, err)

			if errors.Is(e, store.ErrStaleClaim) {
				return true, nil
			}

			if e != nil {
				return true, fmt.Errorf("processing call %d: %w", claimed.ID, e)
			}

			w.logger.WarnContext(ctx, "call retry scheduled",
				"call_id", claimed.ID,
				"attempts", claimed.Attempts,
				"error", err,
			)

			return true, nil

		}
	}

	if err := w.store.CompleteCall(ctx, claimed.ID, claimed.Attempts, outcome); err != nil {
		if errors.Is(err, store.ErrStaleClaim) {
			return true, nil
		}

		return true, fmt.Errorf("completing call %d: %w", claimed.ID, err)
	}

	w.logger.InfoContext(ctx, "call completed", "call_id", claimed.ID, "attempts", claimed.Attempts)

	return true, nil
}
