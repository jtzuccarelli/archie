package call

import "time"

type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusComplete   Status = "complete"
	StatusFailed     Status = "failed"
)

type Call struct {
	ID                  int64
	Filename            string
	Transcript          *string
	Flags               []string
	FlagCount           int
	IsPushy             bool
	Score               int
	CreatedAt           time.Time
	TrackdriveURL       *string
	AgentName           *string
	Disposition         string
	OfferName           string
	AgentTalkTime       int
	ForwardDuration     int
	Status              Status
	Attempts            int
	ProcessingStartedAt *time.Time
	LastError           *string
	NextAttemptAt       time.Time
}
