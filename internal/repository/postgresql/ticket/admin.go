package ticket

import (
	"context"
	"fmt"
	"strings"

	"QuanPhotos/internal/model"
)

// AdminListParams contains parameters for admin listing tickets
type AdminListParams struct {
	Page     int
	PageSize int
	Status   string
	Type     string
	UserID   int64 // Filter by specific user (optional)
}

// AdminListResult contains the result of admin listing tickets
type AdminListResult struct {
	Tickets    []*model.Ticket
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

// AdminList retrieves all tickets for admin view
func (r *TicketRepository) AdminList(ctx context.Context, params AdminListParams) (*AdminListResult, error) {
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

	if params.Status != "" && params.Status != "all" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, params.Status)
		argIndex++
	}

	if params.Type != "" && params.Type != "all" {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, params.Type)
		argIndex++
	}

	if params.UserID > 0 {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, params.UserID)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

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
		ORDER BY
			CASE status
				WHEN 'open' THEN 1
				WHEN 'processing' THEN 2
				WHEN 'resolved' THEN 3
				WHEN 'closed' THEN 4
			END,
			created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, params.PageSize, offset)

	var tickets []*model.Ticket
	err = r.DB().SelectContext(ctx, &tickets, query, args...)
	if err != nil {
		return nil, err
	}

	return &AdminListResult{
		Tickets:    tickets,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// AdminUpdateStatus updates ticket status (for admin)
func (r *TicketRepository) AdminUpdateStatus(ctx context.Context, ticketID int64, status model.TicketStatus) error {
	query := `UPDATE tickets SET status = $1, updated_at = NOW() WHERE id = $2`
	result, err := r.DB().ExecContext(ctx, query, status, ticketID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return nil
	}

	return nil
}

// AdminReply adds a reply to a ticket (for admin, can reply to any ticket)
func (r *TicketRepository) AdminReply(ctx context.Context, ticketID, adminID int64, content string, updateStatus *model.TicketStatus) (int64, error) {
	tx, err := r.DB().BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Create reply
	var replyID int64
	err = tx.QueryRowContext(ctx, `
		INSERT INTO ticket_replies (ticket_id, user_id, content)
		VALUES ($1, $2, $3)
		RETURNING id
	`, ticketID, adminID, content).Scan(&replyID)
	if err != nil {
		return 0, err
	}

	// Update status if provided
	if updateStatus != nil {
		_, err = tx.ExecContext(ctx, `UPDATE tickets SET status = $1, updated_at = NOW() WHERE id = $2`, *updateStatus, ticketID)
		if err != nil {
			return 0, err
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return replyID, nil
}
