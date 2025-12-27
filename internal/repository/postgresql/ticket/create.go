package ticket

import (
	"context"
	"database/sql"

	"QuanPhotos/internal/model"
)

// CreateTicketParams contains parameters for creating a ticket
type CreateTicketParams struct {
	UserID  int64
	PhotoID *int64
	Type    model.TicketType
	Title   string
	Content string
}

// Create creates a new ticket record
func (r *TicketRepository) Create(ctx context.Context, params *CreateTicketParams) (int64, error) {
	query := `
		INSERT INTO tickets (user_id, photo_id, type, title, content, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var id int64
	err := r.DB().QueryRowContext(ctx, query,
		params.UserID,
		toNullInt64(params.PhotoID),
		params.Type,
		params.Title,
		params.Content,
		model.TicketStatusOpen,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

// CreateReply creates a new reply to a ticket
func (r *TicketRepository) CreateReply(ctx context.Context, ticketID, userID int64, content string) (int64, error) {
	query := `
		INSERT INTO ticket_replies (ticket_id, user_id, content)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var id int64
	err := r.DB().QueryRowContext(ctx, query, ticketID, userID, content).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// Helper function for converting pointer to sql.NullInt64
func toNullInt64(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *i, Valid: true}
}
