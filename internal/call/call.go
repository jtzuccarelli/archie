package call

import (
	"time"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusComplete   Status = "complete"
	StatusFailed     Status = "failed"
)

type Call struct {
	ID                  int64
	TdCallID            *string
	AudioFileURL        string
	Transcript          *string
	Flags               []string
	FlagCount           int
	CreatedAt           time.Time
	TrackdriveURL       *string
	AgentName           *string
	Disposition         string
	OfferName           string
	AgentTalkTime       *int
	ForwardDuration     *int
	Status              Status
	Attempts            int
	ProcessingStartedAt *time.Time
	LastError           *string
	NextAttemptAt       time.Time
}

type CallDetail struct {
	ID              int64
	AudioFileURL    string
	Transcript      *string
	Flags           []string
	FlagCount       int
	AgentName       *string
	TrackdriveURL   *string
	Disposition     string
	OfferName       string
	AgentTalkTime   *int
	ForwardDuration *int
}

type CallFilters struct {
	AgentName   string
	Disposition string
	SortBy      string
	SortOrder   string
	DateFrom    *time.Time
	DateTo      *time.Time
	TdCallID    string
	Limit       int
	Offset      int
}

type CallSummary struct {
	ID            int64
	TdCallID      *string
	Status        Status
	CreatedAt     time.Time
	AgentName     *string
	OfferName     string
	Disposition   string
	FlagCount     int
	TrackdriveURL *string
}

type Outcome struct {
	Transcript string
	Flags      []string
}
