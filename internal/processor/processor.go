package processor

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jtzuccarelli/archie/internal/analyzer"
	"github.com/jtzuccarelli/archie/internal/call"
	"github.com/jtzuccarelli/archie/internal/trackdrive"
	"github.com/jtzuccarelli/archie/internal/transcriber"
)

type Processor struct {
	logger      *slog.Logger
	trackdrive  *trackdrive.Client
	transcriber *transcriber.Client
	analyzer    *analyzer.Client
}

func New(td *trackdrive.Client, tr *transcriber.Client, an *analyzer.Client, logger *slog.Logger) *Processor {
	return &Processor{
		logger:      logger,
		trackdrive:  td,
		transcriber: tr,
		analyzer:    an,
	}
}

func (p *Processor) Process(ctx context.Context, c call.Call) (call.Outcome, error) {
	logger := p.logger.With("call_id", c.ID, "attempts", c.Attempts)
	logger.InfoContext(ctx, "downloading call audio")
	audio, err := p.trackdrive.Download(ctx, c.AudioFileURL)
	if err != nil {
		return call.Outcome{}, fmt.Errorf("download: %w", err)
	}

	logger.InfoContext(ctx, "transcribing call audio")
	transcript, err := p.transcriber.Transcribe(ctx, audio)
	if err != nil {
		return call.Outcome{}, fmt.Errorf("transcribe: %w", err)
	}

	logger.InfoContext(ctx, "analyzing call transcript")
	analysis, err := p.analyzer.Analyze(ctx, transcript)
	if err != nil {
		return call.Outcome{}, fmt.Errorf("analyze: %w", err)
	}

	flags := analysis.Flags
	if flags == nil {
		flags = []string{}
	}

	return call.Outcome{
		Transcript: transcript,
		Flags:      flags,
	}, nil
}
