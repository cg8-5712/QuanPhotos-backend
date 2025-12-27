package ticket

import (
	"context"
	"fmt"
	"strings"

	"QuanPhotos/internal/model"
)

// ListParams contains parameters for listing tickets
type ListParams struct {
	UserID   int64
	Status   string
	Type     string
	Page     int
	PageSize int
}

// ListResult contains the result of listing tickets
type ListResult struct {
	Tickets    []*model.Ticket
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

// ListByUserID retrieves a paginated list of tickets for a user
func (r *TicketRepository) ListByUserID(ctx context.Context, params ListParams) (*ListResult, error) {
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

	// User ID is required
	conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
	args = append(args, params.UserID)
	argIndex++

	// Optional status filter
	if params.Status != "" && params.Status != "all" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, params.Status)
		argIndex++
	}

	// Optional type filter
	if params.Type != "" && params.Type != "all" {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, params.Type)
		argIndex++
	}

	whereClause := "WHERE " + strings.Join(conditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tickets %s", whereClause)
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

	// Query tickets
	query := fmt.Sprintf(`
		SELECT * FROM tickets
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, params.PageSize, offset)

	var tickets []*model.Ticket
	err = r.DB().SelectContext(ctx, &tickets, query, args...)
	if err != nil {
		return nil, err
	}

	return &ListResult{
		Tickets:    tickets,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}
