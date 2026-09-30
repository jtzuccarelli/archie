-- +goose NO TRANSACTION
-- +goose Up
ALTER TABLE calls
    ADD COLUMN IF NOT EXISTS td_call_id uuid;

CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_calls_td_call_id
    ON calls (td_call_id);
