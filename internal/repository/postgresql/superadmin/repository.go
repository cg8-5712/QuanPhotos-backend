package superadmin

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"QuanPhotos/internal/model"
	"QuanPhotos/internal/repository/postgresql"

	"github.com/jmoiron/sqlx"
)

// Permission constants
const (
	PermManageAnnouncements = "manage_announcements"
	PermManageFeatured      = "manage_featured"
	PermBanUsers            = "ban_users"
	PermMuteComment         = "mute_comment"
	PermMuteMessage         = "mute_message"
	PermMuteUpload          = "mute_upload"
	PermReviewPhotos        = "review_photos"
	PermDeletePhotos        = "delete_photos"
	PermDeleteComments      = "delete_comments"
	PermManageTickets       = "manage_tickets"
	PermManageCategories    = "manage_categories"
	PermManageTags          = "manage_tags"
	PermViewStatistics      = "view_statistics"
	PermViewUserDetails     = "view_user_details"
)

// AllPermissions is a list of all available permissions
var AllPermissions = []string{
	PermManageAnnouncements,
	PermManageFeatured,
	PermBanUsers,
	PermMuteComment,
	PermMuteMessage,
	PermMuteUpload,
	PermReviewPhotos,
	PermDeletePhotos,
	PermDeleteComments,
	PermManageTickets,
	PermManageCategories,
	PermManageTags,
	PermViewStatistics,
	PermViewUserDetails,
}

// AdminPermission represents an admin permission record
type AdminPermission struct {
	AdminID    int64     `db:"admin_id" json:"admin_id"`
	Permission string    `db:"permission" json:"permission"`
	GrantedBy  int64     `db:"granted_by" json:"granted_by"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

// ReviewerCategory represents a reviewer-category assignment
type ReviewerCategory struct {
	ReviewerID int64     `db:"reviewer_id" json:"reviewer_id"`
	CategoryID int       `db:"category_id" json:"category_id"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

// AdminWithPermissions represents an admin with their permissions
type AdminWithPermissions struct {
	User        *model.User `json:"user"`
	Permissions []string    `json:"permissions"`
}

// ReviewerWithCategories represents a reviewer with their assigned categories
type ReviewerWithCategories struct {
	User       *model.User      `json:"user"`
	Categories []*model.Category `json:"categories"`
}

// SuperadminRepository handles superadmin database operations
type SuperadminRepository struct {
	*postgresql.BaseRepository
}

// NewSuperadminRepository creates a new superadmin repository
func NewSuperadminRepository(db *sqlx.DB) *SuperadminRepository {
	return &SuperadminRepository{
		BaseRepository: postgresql.NewBaseRepository(db),
	}
}

// ListAdmins retrieves all admin users with their permissions
func (r *SuperadminRepository) ListAdmins(ctx context.Context) ([]*AdminWithPermissions, error) {
	// Get all admin users
	var users []*model.User
	err := r.DB().SelectContext(ctx, &users, `
		SELECT * FROM users WHERE role = 'admin' ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}

	result := make([]*AdminWithPermissions, len(users))
	for i, u := range users {
		// Get permissions for this admin
		var permissions []string
		err = r.DB().SelectContext(ctx, &permissions, `
			SELECT permission FROM admin_permissions WHERE admin_id = $1
		`, u.ID)
		if err != nil {
			return nil, err
		}
		result[i] = &AdminWithPermissions{
			User:        u,
			Permissions: permissions,
		}
	}

	return result, nil
}

// GetAdminByID retrieves an admin user by ID
func (r *SuperadminRepository) GetAdminByID(ctx context.Context, adminID int64) (*model.User, error) {
	var user model.User
	err := r.DB().GetContext(ctx, &user, `
		SELECT * FROM users WHERE id = $1 AND role IN ('admin', 'superadmin')
	`, adminID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, postgresql.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetAdminPermissions retrieves all permissions for an admin
func (r *SuperadminRepository) GetAdminPermissions(ctx context.Context, adminID int64) ([]string, error) {
	var permissions []string
	err := r.DB().SelectContext(ctx, &permissions, `
		SELECT permission FROM admin_permissions WHERE admin_id = $1
	`, adminID)
	return permissions, err
}

// GrantPermission grants a permission to an admin
func (r *SuperadminRepository) GrantPermission(ctx context.Context, adminID int64, permission string, grantedBy int64) error {
	_, err := r.DB().ExecContext(ctx, `
		INSERT INTO admin_permissions (admin_id, permission, granted_by)
		VALUES ($1, $2, $3)
		ON CONFLICT (admin_id, permission) DO NOTHING
	`, adminID, permission, grantedBy)
	return err
}

// RevokePermission revokes a permission from an admin
func (r *SuperadminRepository) RevokePermission(ctx context.Context, adminID int64, permission string) error {
	result, err := r.DB().ExecContext(ctx, `
		DELETE FROM admin_permissions WHERE admin_id = $1 AND permission = $2
	`, adminID, permission)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return postgresql.ErrNotFound
	}
	return nil
}

// ListReviewers retrieves all reviewer users with their assigned categories
func (r *SuperadminRepository) ListReviewers(ctx context.Context) ([]*ReviewerWithCategories, error) {
	// Get all reviewer users
	var users []*model.User
	err := r.DB().SelectContext(ctx, &users, `
		SELECT * FROM users WHERE role = 'reviewer' ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}

	result := make([]*ReviewerWithCategories, len(users))
	for i, u := range users {
		// Get assigned categories for this reviewer
		var categories []*model.Category
		err = r.DB().SelectContext(ctx, &categories, `
			SELECT c.* FROM categories c
			INNER JOIN reviewer_categories rc ON c.id = rc.category_id
			WHERE rc.reviewer_id = $1
		`, u.ID)
		if err != nil {
			return nil, err
		}
		result[i] = &ReviewerWithCategories{
			User:       u,
			Categories: categories,
		}
	}

	return result, nil
}

// GetReviewerByID retrieves a reviewer user by ID
func (r *SuperadminRepository) GetReviewerByID(ctx context.Context, reviewerID int64) (*model.User, error) {
	var user model.User
	err := r.DB().GetContext(ctx, &user, `
		SELECT * FROM users WHERE id = $1 AND role = 'reviewer'
	`, reviewerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, postgresql.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetReviewerCategories retrieves all category IDs assigned to a reviewer
func (r *SuperadminRepository) GetReviewerCategories(ctx context.Context, reviewerID int64) ([]int, error) {
	var categoryIDs []int
	err := r.DB().SelectContext(ctx, &categoryIDs, `
		SELECT category_id FROM reviewer_categories WHERE reviewer_id = $1
	`, reviewerID)
	return categoryIDs, err
}

// AssignCategory assigns a category to a reviewer
func (r *SuperadminRepository) AssignCategory(ctx context.Context, reviewerID int64, categoryID int) error {
	_, err := r.DB().ExecContext(ctx, `
		INSERT INTO reviewer_categories (reviewer_id, category_id)
		VALUES ($1, $2)
		ON CONFLICT (reviewer_id, category_id) DO NOTHING
	`, reviewerID, categoryID)
	return err
}

// RevokeCategory revokes a category from a reviewer
func (r *SuperadminRepository) RevokeCategory(ctx context.Context, reviewerID int64, categoryID int) error {
	result, err := r.DB().ExecContext(ctx, `
		DELETE FROM reviewer_categories WHERE reviewer_id = $1 AND category_id = $2
	`, reviewerID, categoryID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return postgresql.ErrNotFound
	}
	return nil
}

// CategoryExists checks if a category exists
func (r *SuperadminRepository) CategoryExists(ctx context.Context, categoryID int) (bool, error) {
	var exists bool
	err := r.DB().GetContext(ctx, &exists, `
		SELECT EXISTS(SELECT 1 FROM categories WHERE id = $1)
	`, categoryID)
	return exists, err
}

// GetUserByID retrieves a user by ID
func (r *SuperadminRepository) GetUserByID(ctx context.Context, userID int64) (*model.User, error) {
	var user model.User
	err := r.DB().GetContext(ctx, &user, `SELECT * FROM users WHERE id = $1`, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, postgresql.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// UpdateUserRestrictions updates user's can_comment, can_message, can_upload flags
func (r *SuperadminRepository) UpdateUserRestrictions(ctx context.Context, userID int64, canComment, canMessage, canUpload *bool) error {
	// Build update query dynamically
	query := "UPDATE users SET updated_at = NOW()"
	args := []interface{}{}
	argIndex := 1

	if canComment != nil {
		query += ", can_comment = $" + string(rune('0'+argIndex))
		args = append(args, *canComment)
		argIndex++
	}
	if canMessage != nil {
		query += ", can_message = $" + string(rune('0'+argIndex))
		args = append(args, *canMessage)
		argIndex++
	}
	if canUpload != nil {
		query += ", can_upload = $" + string(rune('0'+argIndex))
		args = append(args, *canUpload)
		argIndex++
	}

	query += " WHERE id = $" + string(rune('0'+argIndex))
	args = append(args, userID)

	result, err := r.DB().ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return postgresql.ErrNotFound
	}
	return nil
}

// GetUserRestrictions retrieves the restriction status for a user
func (r *SuperadminRepository) GetUserRestrictions(ctx context.Context, userID int64) (canComment, canMessage, canUpload bool, err error) {
	var user struct {
		CanComment bool `db:"can_comment"`
		CanMessage bool `db:"can_message"`
		CanUpload  bool `db:"can_upload"`
	}
	err = r.DB().GetContext(ctx, &user, `
		SELECT can_comment, can_message, can_upload FROM users WHERE id = $1
	`, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, false, false, postgresql.ErrNotFound
		}
		return false, false, false, err
	}
	return user.CanComment, user.CanMessage, user.CanUpload, nil
}
