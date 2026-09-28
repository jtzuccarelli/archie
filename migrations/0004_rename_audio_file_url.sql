-- +goose Up
ALTER TABLE calls
    RENAME COLUMN trackdrive_url TO audio_file_url;

