-- +goose Up
CREATE TABLE users (
    id                 bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email              text NOT NULL,
    password_hash      text NOT NULL,
    session_token_hash text UNIQUE,
    session_expires_at timestamptz
);

CREATE UNIQUE INDEX idx_users_email ON users (lower(email));
