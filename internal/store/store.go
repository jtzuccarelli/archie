package store

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jtzuccarelli/archie/internal/call"
)

var retryDelays = []time.Duration{
	5 * time.Minute,
	15 * time.Minute,
	25 * time.Minute,
}

const staleAfter = 5 * time.Minute

var (
	ErrDuplicateCall = errors.New("call already exists")
	ErrStaleClaim    = errors.New("claim is stale")
)

type Store struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewStore(pool *pgxpool.Pool, logger *slog.Logger) *Store {
	return &Store{pool: pool, logger: logger}
}

func (s *Store) CreateCall(ctx context.Context, c call.Call) (int64, error) {
	s.logger.InfoContext(ctx, "creating call in store", "td_call_id", c.TdCallID)
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
	s.logger.InfoContext(ctx, "claiming calls in store", "limit", limit)
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
	s.logger.InfoContext(ctx, "completing call in store", "call_id", id, "attempts", attempts)

	const query = `
		UPDATE calls
		   SET status = 'complete',
		       transcript = $1,
		       flags = $2,
		       flag_count = jsonb_array_length($2),
		       last_error = NULL
		 WHERE id = $3
		   AND attempts = $4
		   AND status = 'processing'`

	tag, err := s.pool.Exec(ctx, query, o.Transcript, o.Flags, id, attempts)

	if err != nil {
		return fmt.Errorf("completing call %d: %w", id, err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("completing call %d: %w", id, ErrStaleClaim)
	}

	return nil
}

func (s *Store) FailCall(ctx context.Context, id int64, attempts int, errCause error) error {
	s.logger.InfoContext(ctx, "failing call in store", "call_id", id, "attempts", attempts)

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
	s.logger.InfoContext(ctx, "scheduling call retry in store", "call_id", id, "attempts", attempts)

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

	if attempts < 1 || attempts > len(retryDelays) {
		return fmt.Errorf("invalid attempt count: %d", attempts)
	}

	nextAttemptAt := time.Now().Add(retryDelays[attempts-1])

	tag, err := s.pool.Exec(ctx, query, id, attempts, message, nextAttemptAt)

	if err != nil {
		return fmt.Errorf("retrying call %d: %w", id, err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("retrying call %d: %w", id, ErrStaleClaim)
	}

	return nil
}

func (s *Store) RecoverStaleCalls(ctx context.Context) (retried, failed int64, err error) {
	s.logger.InfoContext(ctx, "recovering stale calls in store")

	const query = `
		UPDATE calls
		   SET status = CASE WHEN attempts < 3 THEN 'pending' ELSE 'failed' END,
		       next_attempt_at = CASE
		           WHEN attempts < 3
		           THEN now() + ($2::bigint[])[GREATEST(attempts, 1)] * interval '1 second'
		           ELSE next_attempt_at
		       END,
		       last_error = 'processing lease expired'
		 WHERE status = 'processing'
		   AND processing_started_at < now() - $1::bigint * interval '1 second'
		RETURNING status`

	delaySecs := make([]int64, len(retryDelays))
	for i, d := range retryDelays {
		delaySecs[i] = int64(d.Seconds())
	}

	rows, err := s.pool.Query(ctx, query, int64(staleAfter.Seconds()), delaySecs)
	if err != nil {
		return 0, 0, fmt.Errorf("recovering stale calls: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var status call.Status
		if err := rows.Scan(&status); err != nil {
			return 0, 0, fmt.Errorf("scanning recovered call: %w", err)
		}
		if status == call.StatusFailed {
			failed++
		} else {
			retried++
		}
	}

	if err := rows.Err(); err != nil {
		return 0, 0, fmt.Errorf("iterating recovered calls: %w", err)
	}

	return retried, failed, nil
}

func (s *Store) GetCall(ctx context.Context, id int64) (call.CallDetail, error) {
	s.logger.InfoContext(ctx, "retrieving call", "call_id", id)

	const query = `
		SELECT id, audio_file_url, transcript, flags, flag_count,
		       agent_name, trackdrive_url, disposition, offer_name,
		       agent_talk_time, forward_duration
		  FROM calls
		 WHERE id = $1`

	var detail call.CallDetail
	err := s.pool.QueryRow(ctx, query, id).Scan(
		&detail.ID,
		&detail.AudioFileURL,
		&detail.Transcript,
		&detail.Flags,
		&detail.FlagCount,
		&detail.AgentName,
		&detail.TrackdriveURL,
		&detail.Disposition,
		&detail.OfferName,
		&detail.AgentTalkTime,
		&detail.ForwardDuration,
	)
	if err != nil {
		return call.CallDetail{}, fmt.Errorf("retrieving call %d: %w", id, err)
	}

	return detail, nil
}
