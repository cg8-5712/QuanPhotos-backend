package user

import (
	"context"

	"QuanPhotos/internal/model"

	"github.com/jmoiron/sqlx"
)

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (username, email, password_hash, role, status, can_comment, can_message, can_upload, avatar, bio, location)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`

	return r.DB().QueryRowxContext(ctx, query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.Status,
		user.CanComment,
		user.CanMessage,
		user.CanUpload,
		user.Avatar,
		user.Bio,
		user.Location,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

// CreateTx creates a new user within a transaction
func (r *UserRepository) CreateTx(ctx context.Context, tx *sqlx.Tx, user *model.User) error {
	query := `
		INSERT INTO users (username, email, password_hash, role, status, can_comment, can_message, can_upload, avatar, bio, location)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`

	return tx.QueryRowxContext(ctx, query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.Status,
		user.CanComment,
		user.CanMessage,
		user.CanUpload,
		user.Avatar,
		user.Bio,
		user.Location,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}
