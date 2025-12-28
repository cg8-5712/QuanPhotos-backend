package ticket

import (
	"context"
	"database/sql"
	"errors"

	"QuanPhotos/internal/model"
	"QuanPhotos/internal/repository/postgresql"

	"github.com/jmoiron/sqlx"
)

// GetByID retrieves a ticket by ID
func (r *TicketRepository) GetByID(ctx context.Context, id int64) (*model.Ticket, error) {
	query := `SELECT * FROM tickets WHERE id = $1`

	var ticket model.Ticket
	err := r.DB().GetContext(ctx, &ticket, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, postgresql.ErrNotFound
		}
		return nil, err
	}

	return &ticket, nil
}

// GetReplies retrieves all replies for a ticket
func (r *TicketRepository) GetReplies(ctx context.Context, ticketID int64) ([]*model.TicketReply, error) {
	query := `
		SELECT * FROM ticket_replies
		WHERE ticket_id = $1
		ORDER BY created_at ASC
	`

	var replies []*model.TicketReply
	err := r.DB().SelectContext(ctx, &replies, query, ticketID)
	if err != nil {
		return nil, err
	}

	return replies, nil
}

// GetReplyByID retrieves a single reply by ID
func (r *TicketRepository) GetReplyByID(ctx context.Context, id int64) (*model.TicketReply, error) {
	query := `SELECT * FROM ticket_replies WHERE id = $1`

	var reply model.TicketReply
	err := r.DB().GetContext(ctx, &reply, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, postgresql.ErrNotFound
		}
		return nil, err
	}

	return &reply, nil
}

// Exists checks if a ticket exists
func (r *TicketRepository) Exists(ctx context.Context, id int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM tickets WHERE id = $1)`

	var exists bool
	err := r.DB().GetContext(ctx, &exists, query, id)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// IsOwnedBy checks if a ticket is owned by a user
func (r *TicketRepository) IsOwnedBy(ctx context.Context, ticketID, userID int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM tickets WHERE id = $1 AND user_id = $2)`

	var isOwner bool
	err := r.DB().GetContext(ctx, &isOwner, query, ticketID, userID)
	if err != nil {
		return false, err
	}

	return isOwner, nil
}

// GetPhoto retrieves photo brief info for a ticket (if photo_id exists)
func (r *TicketRepository) GetPhoto(ctx context.Context, photoID int64, baseURL string) (*model.TicketPhotoBrief, error) {
	query := `SELECT id, title, thumbnail_path FROM photos WHERE id = $1`

	var photo struct {
		ID            int64          `db:"id"`
		Title         string         `db:"title"`
		ThumbnailPath sql.NullString `db:"thumbnail_path"`
	}
	err := r.DB().GetContext(ctx, &photo, query, photoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	thumbnailURL := ""
	if photo.ThumbnailPath.Valid {
		thumbnailURL = baseURL + photo.ThumbnailPath.String
	}

	return &model.TicketPhotoBrief{
		ID:           photo.ID,
		Title:        photo.Title,
		ThumbnailURL: thumbnailURL,
	}, nil
}

// GetUserBriefMap retrieves user brief info for a list of user IDs
func (r *TicketRepository) GetUserBriefMap(ctx context.Context, userIDs []int64) (map[int64]*model.TicketUserBrief, error) {
	if len(userIDs) == 0 {
		return make(map[int64]*model.TicketUserBrief), nil
	}

	query, args, err := sqlx.In(`SELECT id, username, role FROM users WHERE id IN (?)`, userIDs)
	if err != nil {
		return nil, err
	}
	query = r.DB().Rebind(query)

	var users []struct {
		ID       int64          `db:"id"`
		Username string         `db:"username"`
		Role     model.UserRole `db:"role"`
	}
	err = r.DB().SelectContext(ctx, &users, query, args...)
	if err != nil {
		return nil, err
	}

	result := make(map[int64]*model.TicketUserBrief)
	for _, u := range users {
		result[u.ID] = &model.TicketUserBrief{
			ID:       u.ID,
			Username: u.Username,
			Role:     u.Role,
		}
	}

	return result, nil
}
