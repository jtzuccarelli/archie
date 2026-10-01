package processor

import (
	"context"
	"fmt"

	"github.com/jtzuccarelli/archie/internal/call"
	"github.com/jtzuccarelli/archie/internal/trackdrive"
)

type Processor struct {
	trackdrive *trackdrive.Client
}

type Result struct {
	Transcript string
	Flags      []string
	FlagCount  int
	IsPushy    bool
	Score      int
}

func New(trackdrive *trackdrive.Client) *Processor {
	return &Processor{
		trackdrive: trackdrive,
	}
}

func (p *Processor) Process(ctx context.Context, c call.Call) (Result, error) {
	audio, err := p.trackdrive.Download(ctx, c.AudioFileURL)
	if err != nil {
		return Result{}, fmt.Errorf("download: %w", err)
	}

	_ = audio

	return Result{}, nil
}
