-- +goose Up
ALTER TABLE calls
    DROP COLUMN is_pushy,
    DROP COLUMN score;
