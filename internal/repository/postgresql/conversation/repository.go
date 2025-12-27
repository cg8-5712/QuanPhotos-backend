package conversation

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"QuanPhotos/internal/model"
	"QuanPhotos/internal/repository/postgresql"

	"github.com/jmoiron/sqlx"
)

// Conversation represents a conversation between two users
type Conversation struct {
	ID            int64          `db:"id" json:"id"`
	User1ID       int64          `db:"user1_id" json:"user1_id"`
	User2ID       int64          `db:"user2_id" json:"user2_id"`
	LastMessageID sql.NullInt64  `db:"last_message_id" json:"-"`
	User1Unread   int            `db:"user1_unread" json:"user1_unread"`
	User2Unread   int            `db:"user2_unread" json:"user2_unread"`
	User1Deleted  bool           `db:"user1_deleted" json:"user1_deleted"`
	User2Deleted  bool           `db:"user2_deleted" json:"user2_deleted"`
	CreatedAt     time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at" json:"updated_at"`
}

// Message represents a message in a conversation
type Message struct {
	ID             int64     `db:"id" json:"id"`
	ConversationID int64     `db:"conversation_id" json:"conversation_id"`
	SenderID       int64     `db:"sender_id" json:"sender_id"`
	Content        string    `db:"content" json:"content"`
	IsRead         bool      `db:"is_read" json:"is_read"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

// ConversationRepository handles conversation database operations
type ConversationRepository struct {
	*postgresql.BaseRepository
}

// NewConversationRepository creates a new conversation repository
func NewConversationRepository(db *sqlx.DB) *ConversationRepository {
	return &ConversationRepository{
		BaseRepository: postgresql.NewBaseRepository(db),
	}
}

// ListParams contains parameters for listing conversations
type ListParams struct {
	UserID   int64
	Page     int
	PageSize int
}

// ListResult contains the result of listing conversations
type ListResult struct {
	Conversations []*Conversation
	Total         int64
	Page          int
	PageSize      int
	TotalPages    int
}

// List retrieves conversations for a user
func (r *ConversationRepository) List(ctx context.Context, params ListParams) (*ListResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}

	// Count total
	countQuery := `
		SELECT COUNT(*) FROM conversations
		WHERE (user1_id = $1 AND user1_deleted = FALSE)
		   OR (user2_id = $1 AND user2_deleted = FALSE)
	`
	var total int64
	err := r.DB().GetContext(ctx, &total, countQuery, params.UserID)
	if err != nil {
		return nil, err
	}

	// Calculate pagination
	offset := (params.Page - 1) * params.PageSize
	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	// Query conversations
	query := `
		SELECT * FROM conversations
		WHERE (user1_id = $1 AND user1_deleted = FALSE)
		   OR (user2_id = $1 AND user2_deleted = FALSE)
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`

	var conversations []*Conversation
	err = r.DB().SelectContext(ctx, &conversations, query, params.UserID, params.PageSize, offset)
	if err != nil {
		return nil, err
	}

	return &ListResult{
		Conversations: conversations,
		Total:         total,
		Page:          params.Page,
		PageSize:      params.PageSize,
		TotalPages:    totalPages,
	}, nil
}

// GetByID retrieves a conversation by ID
func (r *ConversationRepository) GetByID(ctx context.Context, id int64) (*Conversation, error) {
	var conv Conversation
	err := r.DB().GetContext(ctx, &conv, `SELECT * FROM conversations WHERE id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, postgresql.ErrNotFound
		}
		return nil, err
	}
	return &conv, nil
}

// GetOrCreate gets an existing conversation or creates a new one
func (r *ConversationRepository) GetOrCreate(ctx context.Context, userID1, userID2 int64) (*Conversation, error) {
	// Ensure user1_id < user2_id
	if userID1 > userID2 {
		userID1, userID2 = userID2, userID1
	}

	// Try to get existing conversation
	var conv Conversation
	err := r.DB().GetContext(ctx, &conv, `
		SELECT * FROM conversations WHERE user1_id = $1 AND user2_id = $2
	`, userID1, userID2)

	if err == nil {
		// Restore if deleted
		if conv.User1Deleted || conv.User2Deleted {
			_, err = r.DB().ExecContext(ctx, `
				UPDATE conversations SET user1_deleted = FALSE, user2_deleted = FALSE WHERE id = $1
			`, conv.ID)
			if err != nil {
				return nil, err
			}
			conv.User1Deleted = false
			conv.User2Deleted = false
		}
		return &conv, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	// Create new conversation
	err = r.DB().GetContext(ctx, &conv, `
		INSERT INTO conversations (user1_id, user2_id)
		VALUES ($1, $2)
		RETURNING *
	`, userID1, userID2)
	if err != nil {
		return nil, err
	}

	return &conv, nil
}

// IsParticipant checks if a user is a participant in a conversation
func (r *ConversationRepository) IsParticipant(ctx context.Context, convID, userID int64) (bool, error) {
	var exists bool
	err := r.DB().GetContext(ctx, &exists, `
		SELECT EXISTS(
			SELECT 1 FROM conversations
			WHERE id = $1 AND (user1_id = $2 OR user2_id = $2)
		)
	`, convID, userID)
	return exists, err
}

// Delete soft deletes a conversation for a user
func (r *ConversationRepository) Delete(ctx context.Context, convID, userID int64) error {
	// Get conversation
	conv, err := r.GetByID(ctx, convID)
	if err != nil {
		return err
	}

	// Determine which user is deleting
	var updateQuery string
	if conv.User1ID == userID {
		updateQuery = `UPDATE conversations SET user1_deleted = TRUE WHERE id = $1`
	} else if conv.User2ID == userID {
		updateQuery = `UPDATE conversations SET user2_deleted = TRUE WHERE id = $1`
	} else {
		return postgresql.ErrNotFound
	}

	_, err = r.DB().ExecContext(ctx, updateQuery, convID)
	return err
}

// MessageListParams contains parameters for listing messages
type MessageListParams struct {
	ConversationID int64
	Page           int
	PageSize       int
}

// MessageListResult contains the result of listing messages
type MessageListResult struct {
	Messages   []*Message
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
}

// ListMessages retrieves messages in a conversation
func (r *ConversationRepository) ListMessages(ctx context.Context, params MessageListParams) (*MessageListResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 50
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}

	// Count total
	countQuery := `SELECT COUNT(*) FROM messages WHERE conversation_id = $1`
	var total int64
	err := r.DB().GetContext(ctx, &total, countQuery, params.ConversationID)
	if err != nil {
		return nil, err
	}

	// Calculate pagination
	offset := (params.Page - 1) * params.PageSize
	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	// Query messages (newest first)
	query := `
		SELECT * FROM messages
		WHERE conversation_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var messages []*Message
	err = r.DB().SelectContext(ctx, &messages, query, params.ConversationID, params.PageSize, offset)
	if err != nil {
		return nil, err
	}

	return &MessageListResult{
		Messages:   messages,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

// SendMessage sends a message in a conversation
func (r *ConversationRepository) SendMessage(ctx context.Context, convID, senderID int64, content string) (*Message, error) {
	tx, err := r.DB().BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Get conversation to determine recipient
	var conv Conversation
	err = tx.GetContext(ctx, &conv, `SELECT * FROM conversations WHERE id = $1`, convID)
	if err != nil {
		return nil, err
	}

	// Insert message
	var msg Message
	err = tx.GetContext(ctx, &msg, `
		INSERT INTO messages (conversation_id, sender_id, content)
		VALUES ($1, $2, $3)
		RETURNING *
	`, convID, senderID, content)
	if err != nil {
		return nil, err
	}

	// Update conversation
	var updateQuery string
	if conv.User1ID == senderID {
		// Sender is user1, increment user2's unread count
		updateQuery = `
			UPDATE conversations
			SET last_message_id = $1, user2_unread = user2_unread + 1, updated_at = NOW()
			WHERE id = $2
		`
	} else {
		// Sender is user2, increment user1's unread count
		updateQuery = `
			UPDATE conversations
			SET last_message_id = $1, user1_unread = user1_unread + 1, updated_at = NOW()
			WHERE id = $2
		`
	}

	_, err = tx.ExecContext(ctx, updateQuery, msg.ID, convID)
	if err != nil {
		return nil, err
	}

	return &msg, tx.Commit()
}

// MarkAsRead marks all messages in a conversation as read for a user
func (r *ConversationRepository) MarkAsRead(ctx context.Context, convID, userID int64) error {
	tx, err := r.DB().BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Mark messages as read
	_, err = tx.ExecContext(ctx, `
		UPDATE messages SET is_read = TRUE
		WHERE conversation_id = $1 AND sender_id != $2 AND is_read = FALSE
	`, convID, userID)
	if err != nil {
		return err
	}

	// Get conversation
	var conv Conversation
	err = tx.GetContext(ctx, &conv, `SELECT * FROM conversations WHERE id = $1`, convID)
	if err != nil {
		return err
	}

	// Reset unread count
	var updateQuery string
	if conv.User1ID == userID {
		updateQuery = `UPDATE conversations SET user1_unread = 0 WHERE id = $1`
	} else {
		updateQuery = `UPDATE conversations SET user2_unread = 0 WHERE id = $1`
	}

	_, err = tx.ExecContext(ctx, updateQuery, convID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetUnreadCount returns the total unread message count for a user
func (r *ConversationRepository) GetUnreadCount(ctx context.Context, userID int64) (int, error) {
	var count int
	err := r.DB().GetContext(ctx, &count, `
		SELECT COALESCE(
			SUM(CASE WHEN user1_id = $1 THEN user1_unread ELSE user2_unread END),
			0
		)
		FROM conversations
		WHERE (user1_id = $1 AND user1_deleted = FALSE)
		   OR (user2_id = $1 AND user2_deleted = FALSE)
	`, userID)
	return count, err
}

// GetLastMessage retrieves the last message in a conversation
func (r *ConversationRepository) GetLastMessage(ctx context.Context, messageID int64) (*Message, error) {
	var msg Message
	err := r.DB().GetContext(ctx, &msg, `SELECT * FROM messages WHERE id = $1`, messageID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &msg, nil
}

// UserExists checks if a user exists
func (r *ConversationRepository) UserExists(ctx context.Context, userID int64) (bool, error) {
	var exists bool
	err := r.DB().GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND status = 'active')`, userID)
	return exists, err
}

// GetUserMap retrieves a map of users by their IDs
func (r *ConversationRepository) GetUserMap(ctx context.Context, userIDs []int64) (map[int64]*model.User, error) {
	if len(userIDs) == 0 {
		return make(map[int64]*model.User), nil
	}

	query, args, err := sqlx.In(`SELECT * FROM users WHERE id IN (?)`, userIDs)
	if err != nil {
		return nil, err
	}
	query = r.DB().Rebind(query)

	var users []*model.User
	err = r.DB().SelectContext(ctx, &users, query, args...)
	if err != nil {
		return nil, err
	}

	result := make(map[int64]*model.User)
	for _, u := range users {
		result[u.ID] = u
	}

	return result, nil
}
