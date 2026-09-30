package worker

import (
	"context"
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

	return nil
}
