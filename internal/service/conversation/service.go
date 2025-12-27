package conversation

import (
	"context"
	"errors"
	"time"

	"QuanPhotos/internal/repository/postgresql/conversation"
)

var (
	ErrConversationNotFound = errors.New("conversation not found")
	ErrUserNotFound         = errors.New("user not found")
	ErrNotParticipant       = errors.New("you are not a participant of this conversation")
	ErrCannotMessageSelf    = errors.New("cannot send message to yourself")
)

// Service handles conversation business logic
type Service struct {
	convRepo *conversation.ConversationRepository
}

// New creates a new conversation service
func New(convRepo *conversation.ConversationRepository) *Service {
	return &Service{
		convRepo: convRepo,
	}
}

// UserBrief represents brief user info
type UserBrief struct {
	ID       int64   `json:"id"`
	Username string  `json:"username"`
	Avatar   *string `json:"avatar,omitempty"`
}

// MessageItem represents a message in response
type MessageItem struct {
	ID        int64     `json:"id"`
	SenderID  int64     `json:"sender_id"`
	Content   string    `json:"content"`
	IsRead    bool      `json:"is_read"`
	CreatedAt string    `json:"created_at"`
}

// ConversationItem represents a conversation in response
type ConversationItem struct {
	ID           int64        `json:"id"`
	OtherUser    *UserBrief   `json:"other_user"`
	LastMessage  *MessageItem `json:"last_message,omitempty"`
	UnreadCount  int          `json:"unread_count"`
	UpdatedAt    string       `json:"updated_at"`
}

// Pagination represents pagination info
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// ListRequest represents request for listing conversations
type ListRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

// ListResponse represents response for listing conversations
type ListResponse struct {
	List       []ConversationItem `json:"list"`
	Pagination Pagination         `json:"pagination"`
}

// List retrieves conversations for a user
func (s *Service) List(ctx context.Context, userID int64, req *ListRequest) (*ListResponse, error) {
	result, err := s.convRepo.List(ctx, conversation.ListParams{
		UserID:   userID,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}

	// Get unique user IDs
	userIDs := make([]int64, 0, len(result.Conversations)*2)
	userIDMap := make(map[int64]bool)
	for _, c := range result.Conversations {
		if !userIDMap[c.User1ID] {
			userIDs = append(userIDs, c.User1ID)
			userIDMap[c.User1ID] = true
		}
		if !userIDMap[c.User2ID] {
			userIDs = append(userIDs, c.User2ID)
			userIDMap[c.User2ID] = true
		}
	}

	// Get users
	usersModelMap, err := s.convRepo.GetUserMap(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	// Convert to UserBrief map
	usersMap := make(map[int64]*UserBrief)
	for id, u := range usersModelMap {
		ub := &UserBrief{
			ID:       u.ID,
			Username: u.Username,
		}
		if u.Avatar.Valid {
			ub.Avatar = &u.Avatar.String
		}
		usersMap[id] = ub
	}

	// Build response
	list := make([]ConversationItem, len(result.Conversations))
	for i, c := range result.Conversations {
		// Determine other user and unread count
		var otherUserID int64
		var unreadCount int
		if c.User1ID == userID {
			otherUserID = c.User2ID
			unreadCount = c.User1Unread
		} else {
			otherUserID = c.User1ID
			unreadCount = c.User2Unread
		}

		item := ConversationItem{
			ID:          c.ID,
			OtherUser:   usersMap[otherUserID],
			UnreadCount: unreadCount,
			UpdatedAt:   c.UpdatedAt.Format(time.RFC3339),
		}

		// Get last message
		if c.LastMessageID.Valid {
			msg, err := s.convRepo.GetLastMessage(ctx, c.LastMessageID.Int64)
			if err == nil && msg != nil {
				item.LastMessage = &MessageItem{
					ID:        msg.ID,
					SenderID:  msg.SenderID,
					Content:   msg.Content,
					IsRead:    msg.IsRead,
					CreatedAt: msg.CreatedAt.Format(time.RFC3339),
				}
			}
		}

		list[i] = item
	}

	return &ListResponse{
		List: list,
		Pagination: Pagination{
			Page:       result.Page,
			PageSize:   result.PageSize,
			Total:      result.Total,
			TotalPages: result.TotalPages,
		},
	}, nil
}

// CreateRequest represents request for creating a conversation
type CreateRequest struct {
	RecipientID int64  `json:"recipient_id" binding:"required"`
	Content     string `json:"content" binding:"required,min=1,max=2000"`
}

// CreateResponse represents response for creating a conversation
type CreateResponse struct {
	ConversationID int64        `json:"conversation_id"`
	Message        *MessageItem `json:"message"`
}

// Create creates a conversation and sends the first message
func (s *Service) Create(ctx context.Context, senderID int64, req *CreateRequest) (*CreateResponse, error) {
	// Check if sending to self
	if senderID == req.RecipientID {
		return nil, ErrCannotMessageSelf
	}

	// Check if recipient exists
	exists, err := s.convRepo.UserExists(ctx, req.RecipientID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrUserNotFound
	}

	// Get or create conversation
	conv, err := s.convRepo.GetOrCreate(ctx, senderID, req.RecipientID)
	if err != nil {
		return nil, err
	}

	// Send message
	msg, err := s.convRepo.SendMessage(ctx, conv.ID, senderID, req.Content)
	if err != nil {
		return nil, err
	}

	return &CreateResponse{
		ConversationID: conv.ID,
		Message: &MessageItem{
			ID:        msg.ID,
			SenderID:  msg.SenderID,
			Content:   msg.Content,
			IsRead:    msg.IsRead,
			CreatedAt: msg.CreatedAt.Format(time.RFC3339),
		},
	}, nil
}

// GetMessagesRequest represents request for getting messages
type GetMessagesRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

// GetMessagesResponse represents response for getting messages
type GetMessagesResponse struct {
	List       []MessageItem `json:"list"`
	Pagination Pagination    `json:"pagination"`
}

// GetMessages retrieves messages in a conversation
func (s *Service) GetMessages(ctx context.Context, userID, convID int64, req *GetMessagesRequest) (*GetMessagesResponse, error) {
	// Check if participant
	isParticipant, err := s.convRepo.IsParticipant(ctx, convID, userID)
	if err != nil {
		return nil, err
	}
	if !isParticipant {
		return nil, ErrNotParticipant
	}

	// Mark as read
	err = s.convRepo.MarkAsRead(ctx, convID, userID)
	if err != nil {
		// Log but don't fail
	}

	// Get messages
	result, err := s.convRepo.ListMessages(ctx, conversation.MessageListParams{
		ConversationID: convID,
		Page:           req.Page,
		PageSize:       req.PageSize,
	})
	if err != nil {
		return nil, err
	}

	// Build response
	list := make([]MessageItem, len(result.Messages))
	for i, m := range result.Messages {
		list[i] = MessageItem{
			ID:        m.ID,
			SenderID:  m.SenderID,
			Content:   m.Content,
			IsRead:    m.IsRead,
			CreatedAt: m.CreatedAt.Format(time.RFC3339),
		}
	}

	return &GetMessagesResponse{
		List: list,
		Pagination: Pagination{
			Page:       result.Page,
			PageSize:   result.PageSize,
			Total:      result.Total,
			TotalPages: result.TotalPages,
		},
	}, nil
}

// SendMessageRequest represents request for sending a message
type SendMessageRequest struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
}

// SendMessage sends a message in a conversation
func (s *Service) SendMessage(ctx context.Context, userID, convID int64, req *SendMessageRequest) (*MessageItem, error) {
	// Check if participant
	isParticipant, err := s.convRepo.IsParticipant(ctx, convID, userID)
	if err != nil {
		return nil, err
	}
	if !isParticipant {
		return nil, ErrNotParticipant
	}

	// Send message
	msg, err := s.convRepo.SendMessage(ctx, convID, userID, req.Content)
	if err != nil {
		return nil, err
	}

	return &MessageItem{
		ID:        msg.ID,
		SenderID:  msg.SenderID,
		Content:   msg.Content,
		IsRead:    msg.IsRead,
		CreatedAt: msg.CreatedAt.Format(time.RFC3339),
	}, nil
}

// Delete deletes a conversation for a user
func (s *Service) Delete(ctx context.Context, userID, convID int64) error {
	// Check if participant
	isParticipant, err := s.convRepo.IsParticipant(ctx, convID, userID)
	if err != nil {
		return err
	}
	if !isParticipant {
		return ErrNotParticipant
	}

	return s.convRepo.Delete(ctx, convID, userID)
}

// GetUnreadCount returns the total unread message count for a user
func (s *Service) GetUnreadCount(ctx context.Context, userID int64) (int, error) {
	return s.convRepo.GetUnreadCount(ctx, userID)
}
