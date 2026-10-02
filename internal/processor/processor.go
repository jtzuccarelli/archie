package processor

import (
	"context"
	"fmt"

	"github.com/jtzuccarelli/archie/internal/analyzer"
	"github.com/jtzuccarelli/archie/internal/call"
	"github.com/jtzuccarelli/archie/internal/trackdrive"
	"github.com/jtzuccarelli/archie/internal/transcriber"
)

type Processor struct {
	trackdrive  *trackdrive.Client
	transcriber *transcriber.Client
	analyzer    *analyzer.Client
}

type Result struct {
	Transcript string
	Flags      []string
	FlagCount  int
	IsPushy    bool
	Score      int
}

func New(td *trackdrive.Client, tr *transcriber.Client, an *analyzer.Client) *Processor {
	return &Processor{
		trackdrive:  td,
		transcriber: tr,
		analyzer:    an,
	}
}

func (p *Processor) Process(ctx context.Context, c call.Call) (Result, error) {
	audio, err := p.trackdrive.Download(ctx, c.AudioFileURL)
	if err != nil {
		return Result{}, fmt.Errorf("download: %w", err)
	}

	transcript, err := p.transcriber.Transcribe(ctx, audio)
	if err != nil {
		return Result{}, fmt.Errorf("transcribe: %w", err)
	}

	analysis, err := p.analyzer.Analyze(ctx, transcript)
	if err != nil {
		return Result{}, fmt.Errorf("analyze: %w", err)
	}

	return Result{
		Transcript: transcript,
		Flags:      analysis.Flags,
		FlagCount:  len(analysis.Flags),
		IsPushy:    analysis.IsPushy,
		Score:      analysis.Score,
	}, nil
}
