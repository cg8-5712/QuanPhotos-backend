package user

import (
	"context"
	"fmt"
	"strings"

	"QuanPhotos/internal/model"
)

// ListParams contains parameters for listing users
type ListParams struct {
	Page     int
	PageSize int
	Role     string
	Status   string
	Keyword  string
}

// ListResult contains the result of listing users
type ListResult struct {
	Users      []*model.User
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

// List retrieves a paginated list of users with optional filters
func (r *UserRepository) List(ctx context.Context, params ListParams) (*ListResult, error) {
	// Set defaults
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}

	// Build WHERE clause
	var conditions []string
	var args []interface{}
	argIndex := 1

	if params.Role != "" && params.Role != "all" {
		conditions = append(conditions, fmt.Sprintf("role = $%d", argIndex))
		args = append(args, params.Role)
		argIndex++
	}

	if params.Status != "" && params.Status != "all" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, params.Status)
		argIndex++
	}

	if params.Keyword != "" {
		conditions = append(conditions, fmt.Sprintf("(username ILIKE $%d OR email ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+params.Keyword+"%")
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users %s", whereClause)
	var total int64
	err := r.DB().GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, err
	}

	// Calculate pagination
	offset := (params.Page - 1) * params.PageSize
	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	// Query users
	query := fmt.Sprintf(`
		SELECT * FROM users
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, params.PageSize, offset)

	var users []*model.User
	err = r.DB().SelectContext(ctx, &users, query, args...)
	if err != nil {
		return nil, err
	}

	return &ListResult{
		Users:      users,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}
