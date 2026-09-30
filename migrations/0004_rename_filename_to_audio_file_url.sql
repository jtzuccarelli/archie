-- +goose Up
ALTER TABLE calls
    RENAME COLUMN filename TO audio_file_url;
