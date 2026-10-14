package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jtzuccarelli/archie/internal/call"
)

var (
	ErrDuplicateCall = errors.New("call already exists")
	ErrStaleClaim    = errors.New("claim is stale")
)

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

func (s *Store) ClaimCalls(ctx context.Context, limit int) ([]call.Call, error) {
	const query = `
		UPDATE calls
		   SET status = 'processing',
		       processing_started_at = now(),
		       attempts = attempts + 1
		 WHERE id IN (
			SELECT id
			  FROM calls
			 WHERE status = 'pending'
			   AND attempts < 3
			   AND next_attempt_at <= now()
			 ORDER BY next_attempt_at
			 LIMIT $1
			   FOR UPDATE SKIP LOCKED
		 )
		RETURNING id, td_call_id, audio_file_url, status, attempts, processing_started_at`

	rows, err := s.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("claiming calls: %w", err)
	}
	defer rows.Close()

	var calls []call.Call
	for rows.Next() {
		var c call.Call
		if err := rows.Scan(
			&c.ID,
			&c.TdCallID,
			&c.AudioFileURL,
			&c.Status,
			&c.Attempts,
			&c.ProcessingStartedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning claimed call: %w", err)
		}
		calls = append(calls, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating claimed calls: %w", err)
	}

	return calls, nil
}

func (s *Store) CompleteCall(ctx context.Context, id int64, attempts int, o call.Outcome) error {

	const query = `
		UPDATE calls
		   SET status = 'complete',
		       transcript = $1,
		       flags = $2,
		       flag_count = jsonb_array_length($2),
		       is_pushy = $3,
		       score = $4,
		       last_error = NULL
		 WHERE id = $5
		   AND attempts = $6
		   AND status = 'processing'`

	tag, err := s.pool.Exec(ctx, query, o.Transcript, o.Flags, o.IsPushy, o.Score, id, attempts)

	if err != nil {
		return fmt.Errorf("completing call %d: %w", id, err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("completing call %d: %w", id, ErrStaleClaim)
	}

	return nil
}

func (s *Store) FailCall(ctx context.Context, id int64, attempts int, errCause error) error {

	const query = `
		UPDATE calls
		   SET status = 'failed',
		       last_error = $3
		 WHERE id = $1
		   AND attempts = $2
		   AND status = 'processing'`

	message := "unknown processing error"
	if errCause != nil {
		message = errCause.Error()
	}

	tag, err := s.pool.Exec(ctx, query, id, attempts, message)

	if err != nil {
		return fmt.Errorf("failing call %d: %w", id, err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("failing call %d: %w", id, ErrStaleClaim)
	}

	return nil
}

func (s *Store) RetryCall(ctx context.Context, id int64, attempts int, errCause error) error {

	const query = `
		UPDATE calls
		   SET status = 'pending',
		       last_error = $3,
			   next_attempt_at = $4
		 WHERE id = $1
		   AND attempts = $2
		   AND status = 'processing'`

	message := "unknown processing error"
	if errCause != nil {
		message = errCause.Error()
	}

	delays := []time.Duration{
		5 * time.Minute,
		15 * time.Minute,
		25 * time.Minute,
	}

	if attempts < 1 || attempts > len(delays) {
		return fmt.Errorf("invalid attempt count: %d", attempts)
	}

	nextAttemptAt := time.Now().Add(delays[attempts-1])

	tag, err := s.pool.Exec(ctx, query, id, attempts, message, nextAttemptAt)

	if err != nil {
		return fmt.Errorf("retrying call %d: %w", id, err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("retrying call %d: %w", id, ErrStaleClaim)
	}

	return nil
}
