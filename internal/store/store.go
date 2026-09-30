package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jtzuccarelli/archie/internal/call"
)

var ErrDuplicateCall = errors.New("call already exists")

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) CreateCall(ctx context.Context, c call.Call) (int64, error) {
	const query = `
		INSERT INTO calls (
			td_call_id,
			audio_file_url,
			trackdrive_url,
			agent_name,
			offer_name,
			disposition,
			agent_talk_time,
			forward_duration,
			status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (td_call_id) DO NOTHING
		RETURNING id`

	var id int64
	err := s.pool.QueryRow(ctx, query,
		c.TdCallID,
		c.AudioFileURL,
		c.TrackdriveURL,
		c.AgentName,
		c.OfferName,
		c.Disposition,
		c.AgentTalkTime,
		c.ForwardDuration,
		call.StatusPending,
	).Scan(&id)

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrDuplicateCall
	}
	if err != nil {
		return 0, fmt.Errorf("inserting call: %w", err)
	}

	return id, nil
}
