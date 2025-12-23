package user

import (
	"context"
	"database/sql"
	"errors"

	"QuanPhotos/internal/model"
	"QuanPhotos/internal/repository/postgresql"
)

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users SET
			username = $1,
			email = $2,
			role = $3,
			status = $4,
			can_comment = $5,
			can_message = $6,
			can_upload = $7,
			avatar = $8,
			bio = $9,
			location = $10,
			updated_at = NOW()
		WHERE id = $11
		RETURNING updated_at
	`

	err := r.DB().QueryRowxContext(ctx, query,
		user.Username,
		user.Email,
		user.Role,
		user.Status,
		user.CanComment,
		user.CanMessage,
		user.CanUpload,
		user.Avatar,
		user.Bio,
		user.Location,
		user.ID,
	).Scan(&user.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return postgresql.ErrNotFound
		}
		return err
	}

	return nil
}

// UpdatePassword updates user's password
func (r *UserRepository) UpdatePassword(ctx context.Context, userID int64, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`

	result, err := r.DB().ExecContext(ctx, query, passwordHash, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return postgresql.ErrNotFound
	}

	return nil
}

// UpdateLastLogin updates user's last login time
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID int64) error {
	query := `UPDATE users SET last_login_at = NOW() WHERE id = $1`
	_, err := r.DB().ExecContext(ctx, query, userID)
	return err
}

// UpdateProfile updates user's profile information
func (r *UserRepository) UpdateProfile(ctx context.Context, userID int64, avatar, bio, location *string) error {
	query := `
		UPDATE users SET
			avatar = COALESCE($1, avatar),
			bio = COALESCE($2, bio),
			location = COALESCE($3, location),
			updated_at = NOW()
		WHERE id = $4
	`

	result, err := r.DB().ExecContext(ctx, query, avatar, bio, location, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return postgresql.ErrNotFound
	}

	return nil
}

// UpdateRole updates user's role
func (r *UserRepository) UpdateRole(ctx context.Context, userID int64, role model.UserRole) error {
	query := `UPDATE users SET role = $1, updated_at = NOW() WHERE id = $2`

	result, err := r.DB().ExecContext(ctx, query, role, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return postgresql.ErrUserNotFound
	}

	return nil
}

// UpdateStatus updates user's status
func (r *UserRepository) UpdateStatus(ctx context.Context, userID int64, status model.UserStatus) error {
	query := `UPDATE users SET status = $1, updated_at = NOW() WHERE id = $2`

	result, err := r.DB().ExecContext(ctx, query, status, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return postgresql.ErrUserNotFound
	}

	return nil
}

// UpdateRestrictions updates user's restrictions
func (r *UserRepository) UpdateRestrictions(ctx context.Context, userID int64, canComment, canMessage, canUpload bool) error {
	query := `
		UPDATE users SET
			can_comment = $1,
			can_message = $2,
			can_upload = $3,
			updated_at = NOW()
		WHERE id = $4
	`

	result, err := r.DB().ExecContext(ctx, query, canComment, canMessage, canUpload, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return postgresql.ErrUserNotFound
	}

	return nil
}
