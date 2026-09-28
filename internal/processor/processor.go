package processor

import "github.com/jtzuccarelli/archie/internal/trackdrive"

type Processor struct {
	trackdrive *trackdrive.Client
}

func New(trackdrive *trackdrive.Client) *Processor {
	return &Processor{
		trackdrive: trackdrive,
	}
}
