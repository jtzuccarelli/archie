BEGIN;

ALTER TABLE calls
    ADD COLUMN next_attempt_at timestamptz NOT NULL DEFAULT now();

DROP INDEX idx_calls_pending;
CREATE INDEX idx_calls_pending_due ON calls (next_attempt_at) WHERE status = 'pending';

DROP INDEX idx_calls_processing;
CREATE INDEX idx_calls_processing ON calls (processing_started_at) WHERE status = 'processing';

COMMIT;