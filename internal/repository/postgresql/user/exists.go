package user

import (
	"context"
)

// ExistsByUsername checks if a username exists
func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`

	err := r.DB().GetContext(ctx, &exists, query, username)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// ExistsByEmail checks if an email exists
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	err := r.DB().GetContext(ctx, &exists, query, email)
	if err != nil {
		return false, err
	}

	return exists, nil
}
