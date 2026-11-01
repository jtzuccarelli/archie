package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jtzuccarelli/archie/internal/user"
)

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	const query = `SELECT id, email, password_hash FROM users WHERE email = $1`
	var u user.User
	err := s.pool.QueryRow(ctx, query, email).Scan(&u.ID, &u.Email, &u.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("retrieving user %s: %w", email, err)
	}

	return &u, nil
}

func (s *Store) SaveUserSession(ctx context.Context, id int64, tokenHash string, expiresAt time.Time) error {
	const query = `
		UPDATE users
		   SET session_token_hash = $1,
		       session_expires_at = $2
		 WHERE id = $3`

	tag, err := s.pool.Exec(ctx, query, tokenHash, expiresAt, id)
	if err != nil {
		return fmt.Errorf("saving session for user %d: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("saving session for user %d: %w", id, pgx.ErrNoRows)
	}
	return nil
}

func (s *Store) HasValidSession(ctx context.Context, tokenHash string) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1 FROM users
			 WHERE session_token_hash = $1
			   AND session_expires_at > now()
		)`

	var valid bool
	if err := s.pool.QueryRow(ctx, query, tokenHash).Scan(&valid); err != nil {
		return false, fmt.Errorf("checking user session: %w", err)
	}
	return valid, nil
}
