package ticket

import (
	"context"
	"errors"

	"QuanPhotos/internal/model"
	"QuanPhotos/internal/repository/postgresql"
	"QuanPhotos/internal/repository/postgresql/ticket"
)

var (
	ErrTicketNotFound = errors.New("ticket not found")
	ErrNotOwner       = errors.New("you are not the owner of this ticket")
	ErrTicketClosed   = errors.New("ticket is closed")
)

// Service handles ticket business logic
type Service struct {
	ticketRepo *ticket.TicketRepository
	baseURL    string
}

// New creates a new ticket service
func New(ticketRepo *ticket.TicketRepository, baseURL string) *Service {
	return &Service{
		ticketRepo: ticketRepo,
		baseURL:    baseURL,
	}
}

// CreateRequest represents request for creating a ticket
type CreateRequest struct {
	PhotoID *int64           `json:"photo_id"`
	Type    model.TicketType `json:"type" binding:"required,oneof=appeal report other"`
	Title   string           `json:"title" binding:"required,max=200"`
	Content string           `json:"content" binding:"required"`
}

// CreateResponse represents response for creating a ticket
type CreateResponse struct {
	ID        int64  `json:"id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// Create creates a new ticket
func (s *Service) Create(ctx context.Context, userID int64, req *CreateRequest) (*CreateResponse, error) {
	id, err := s.ticketRepo.Create(ctx, &ticket.CreateTicketParams{
		UserID:  userID,
		PhotoID: req.PhotoID,
		Type:    req.Type,
		Title:   req.Title,
		Content: req.Content,
	})
	if err != nil {
		return nil, err
	}

	// Get created ticket
	t, err := s.ticketRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &CreateResponse{
		ID:        t.ID,
		Status:    string(t.Status),
		CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// ListRequest represents request for listing tickets
type ListRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Status   string `form:"status"`
	Type     string `form:"type"`
}

// ListResponse represents response for listing tickets
type ListResponse struct {
	List       []*model.TicketListItem `json:"list"`
	Pagination Pagination              `json:"pagination"`
}

// Pagination represents pagination info
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// List retrieves a paginated list of tickets for a user
func (s *Service) List(ctx context.Context, userID int64, req *ListRequest) (*ListResponse, error) {
	result, err := s.ticketRepo.ListByUserID(ctx, ticket.ListParams{
		UserID:   userID,
		Status:   req.Status,
		Type:     req.Type,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}

	// Build response
	list := make([]*model.TicketListItem, len(result.Tickets))
	for i, t := range result.Tickets {
		list[i] = t.ToListItem()
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

// GetDetail retrieves ticket detail with replies
func (s *Service) GetDetail(ctx context.Context, userID, ticketID int64) (*model.TicketDetail, error) {
	// Get ticket
	t, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		if errors.Is(err, postgresql.ErrNotFound) {
			return nil, ErrTicketNotFound
		}
		return nil, err
	}

	// Check ownership
	if t.UserID != userID {
		return nil, ErrNotOwner
	}

	// Get photo info if exists
	var photoBrief *model.TicketPhotoBrief
	if t.PhotoID.Valid {
		photoBrief, _ = s.ticketRepo.GetPhoto(ctx, t.PhotoID.Int64, s.baseURL)
	}

	// Get replies
	replies, err := s.ticketRepo.GetReplies(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	// Get user IDs for replies
	userIDs := make([]int64, 0, len(replies))
	for _, r := range replies {
		userIDs = append(userIDs, r.UserID)
	}

	// Get users map
	usersMap, err := s.ticketRepo.GetUserBriefMap(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	// Build reply items
	replyItems := make([]*model.TicketReplyItem, len(replies))
	for i, r := range replies {
		replyItems[i] = r.ToReplyItem(usersMap[r.UserID])
	}

	return t.ToDetail(photoBrief, replyItems), nil
}

// ReplyRequest represents request for replying to a ticket
type ReplyRequest struct {
	Content string `json:"content" binding:"required"`
}

// ReplyResponse represents response for replying to a ticket
type ReplyResponse struct {
	ID        int64  `json:"id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// Reply adds a reply to a ticket
func (s *Service) Reply(ctx context.Context, userID, ticketID int64, req *ReplyRequest) (*ReplyResponse, error) {
	// Get ticket to check ownership and status
	t, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		if errors.Is(err, postgresql.ErrNotFound) {
			return nil, ErrTicketNotFound
		}
		return nil, err
	}

	// Check ownership
	if t.UserID != userID {
		return nil, ErrNotOwner
	}

	// Check if ticket is closed
	if t.Status == model.TicketStatusClosed {
		return nil, ErrTicketClosed
	}

	// Create reply
	replyID, err := s.ticketRepo.CreateReply(ctx, ticketID, userID, req.Content)
	if err != nil {
		return nil, err
	}

	// Get created reply
	reply, err := s.ticketRepo.GetReplyByID(ctx, replyID)
	if err != nil {
		return nil, err
	}

	return &ReplyResponse{
		ID:        reply.ID,
		Content:   reply.Content,
		CreatedAt: reply.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
