package worker

import (
	"context"
	"fmt"
	"log/slog"

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

	c, err := w.store.ClaimCalls(ctx, 1)

	if err != nil {
		return err
	}

	if len(c) == 0 {
		return nil
	}

	claimed := c[0]

	outcome, err := w.processor.Process(ctx, claimed)

	if err != nil {
		return fmt.Errorf("processing call %d: %w", claimed.ID, err)
	}

	if err = w.store.CompleteCall(ctx, claimed.ID, claimed.Attempts, outcome); err != nil {
		return err
	}

	return nil
}
