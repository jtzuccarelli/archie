package store

import (
	"context"
	"fmt"

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
