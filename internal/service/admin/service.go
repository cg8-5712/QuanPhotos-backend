package admin

import (
	"context"
	"errors"
	"time"

	"QuanPhotos/internal/model"
	"QuanPhotos/internal/repository/postgresql"
	"QuanPhotos/internal/repository/postgresql/photo"
	"QuanPhotos/internal/repository/postgresql/ticket"
	"QuanPhotos/internal/repository/postgresql/user"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrCannotChangeSelf   = errors.New("cannot change own role or status")
	ErrInvalidRole        = errors.New("invalid role")
	ErrInsufficientPerm   = errors.New("insufficient permissions")
	ErrPhotoNotFound      = errors.New("photo not found")
	ErrTicketNotFound     = errors.New("ticket not found")
	ErrAnnouncementNotFound = errors.New("announcement not found")
	ErrAlreadyFeatured    = errors.New("photo is already featured")
	ErrNotFeatured        = errors.New("photo is not featured")
)

// Service handles admin business logic
type Service struct {
	userRepo   *user.UserRepository
	photoRepo  *photo.PhotoRepository
	ticketRepo *ticket.TicketRepository
	baseURL    string
}

// New creates a new admin service
func New(userRepo *user.UserRepository) *Service {
	return &Service{
		userRepo: userRepo,
	}
}

// NewFull creates a new admin service with all dependencies
func NewFull(userRepo *user.UserRepository, photoRepo *photo.PhotoRepository, ticketRepo *ticket.TicketRepository, baseURL string) *Service {
	return &Service{
		userRepo:   userRepo,
		photoRepo:  photoRepo,
		ticketRepo: ticketRepo,
		baseURL:    baseURL,
	}
}

// ListUsersRequest represents request for listing users
type ListUsersRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Role     string `form:"role"`
	Status   string `form:"status"`
	Keyword  string `form:"keyword"`
}

// UserListItem represents a user in the admin list
type UserListItem struct {
	ID          int64   `json:"id"`
	Username    string  `json:"username"`
	Email       string  `json:"email"`
	Role        string  `json:"role"`
	Status      string  `json:"status"`
	Avatar      *string `json:"avatar"`
	PhotoCount  int     `json:"photo_count"`
	CreatedAt   string  `json:"created_at"`
	LastLoginAt *string `json:"last_login_at"`
}

// ListUsersResponse represents response for listing users
type ListUsersResponse struct {
	List       []UserListItem `json:"list"`
	Pagination Pagination     `json:"pagination"`
}

// Pagination represents pagination info
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// ListUsers retrieves a paginated list of users
func (s *Service) ListUsers(ctx context.Context, req *ListUsersRequest) (*ListUsersResponse, error) {
	result, err := s.userRepo.List(ctx, user.ListParams{
		Page:     req.Page,
		PageSize: req.PageSize,
		Role:     req.Role,
		Status:   req.Status,
		Keyword:  req.Keyword,
	})
	if err != nil {
		return nil, err
	}

	list := make([]UserListItem, len(result.Users))
	for i, u := range result.Users {
		item := UserListItem{
			ID:        u.ID,
			Username:  u.Username,
			Email:     u.Email,
			Role:      string(u.Role),
			Status:    string(u.Status),
			CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if u.Avatar.Valid {
			item.Avatar = &u.Avatar.String
		}
		if u.LastLoginAt.Valid {
			lastLogin := u.LastLoginAt.Time.Format("2006-01-02T15:04:05Z07:00")
			item.LastLoginAt = &lastLogin
		}
		list[i] = item
	}

	return &ListUsersResponse{
		List: list,
		Pagination: Pagination{
			Page:       result.Page,
			PageSize:   result.PageSize,
			Total:      result.Total,
			TotalPages: result.TotalPages,
		},
	}, nil
}

// UpdateRoleRequest represents request for updating user role
type UpdateRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=user reviewer admin"`
}

// UpdateUserRole updates a user's role
func (s *Service) UpdateUserRole(ctx context.Context, operatorID, targetUserID int64, operatorRole model.UserRole, req *UpdateRoleRequest) error {
	// Cannot change own role
	if operatorID == targetUserID {
		return ErrCannotChangeSelf
	}

	// Get target user
	targetUser, err := s.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		if errors.Is(err, postgresql.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	newRole := model.UserRole(req.Role)

	// Check permissions
	// Admin can only change users and reviewers, not other admins
	if operatorRole == model.RoleAdmin {
		if targetUser.Role == model.RoleAdmin || targetUser.Role == model.RoleSuperAdmin {
			return ErrInsufficientPerm
		}
		if newRole == model.RoleAdmin {
			return ErrInsufficientPerm
		}
	}

	// Cannot set anyone to superadmin via this endpoint
	if newRole == model.RoleSuperAdmin {
		return ErrInvalidRole
	}

	return s.userRepo.UpdateRole(ctx, targetUserID, newRole)
}

// UpdateStatusRequest represents request for updating user status
type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active banned"`
	Reason string `json:"reason"`
}

// UpdateUserStatus updates a user's status (ban/unban)
func (s *Service) UpdateUserStatus(ctx context.Context, operatorID, targetUserID int64, operatorRole model.UserRole, req *UpdateStatusRequest) error {
	// Cannot change own status
	if operatorID == targetUserID {
		return ErrCannotChangeSelf
	}

	// Get target user
	targetUser, err := s.userRepo.GetByID(ctx, targetUserID)
	if err != nil {
		if errors.Is(err, postgresql.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	// Admin cannot ban other admins or superadmins
	if operatorRole == model.RoleAdmin {
		if targetUser.Role == model.RoleAdmin || targetUser.Role == model.RoleSuperAdmin {
			return ErrInsufficientPerm
		}
	}

	// Superadmin can ban anyone except other superadmins
	if operatorRole == model.RoleSuperAdmin {
		if targetUser.Role == model.RoleSuperAdmin {
			return ErrInsufficientPerm
		}
	}

	return s.userRepo.UpdateStatus(ctx, targetUserID, model.UserStatus(req.Status))
}

// GetUser retrieves a user by ID for admin view
func (s *Service) GetUser(ctx context.Context, userID int64) (*model.User, error) {
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, postgresql.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}

// ============================================
// Photo Review Methods
// ============================================

// ListReviewsRequest represents request for listing pending reviews
type ListReviewsRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Status   string `form:"status"` // ai_passed, ai_rejected, pending, all
}

// ReviewListItem represents a photo in review list
type ReviewListItem struct {
	ID           int64   `json:"id"`
	Title        string  `json:"title"`
	ThumbnailURL string  `json:"thumbnail_url"`
	Status       string  `json:"status"`
	UserID       int64   `json:"user_id"`
	Username     string  `json:"username"`
	CreatedAt    string  `json:"created_at"`
}

// ListReviewsResponse represents response for listing reviews
type ListReviewsResponse struct {
	List       []ReviewListItem `json:"list"`
	Pagination Pagination       `json:"pagination"`
}

// ListReviews retrieves photos pending review
func (s *Service) ListReviews(ctx context.Context, req *ListReviewsRequest) (*ListReviewsResponse, error) {
	result, err := s.photoRepo.ListPendingReviews(ctx, photo.ReviewListParams{
		Page:     req.Page,
		PageSize: req.PageSize,
		Status:   req.Status,
	})
	if err != nil {
		return nil, err
	}

	// Get user IDs
	userIDs := make([]int64, 0, len(result.Photos))
	userIDMap := make(map[int64]bool)
	for _, p := range result.Photos {
		if !userIDMap[p.UserID] {
			userIDs = append(userIDs, p.UserID)
			userIDMap[p.UserID] = true
		}
	}

	// Get users
	users, err := s.photoRepo.GetUserMap(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	list := make([]ReviewListItem, len(result.Photos))
	for i, p := range result.Photos {
		item := ReviewListItem{
			ID:        p.ID,
			Title:     p.Title,
			Status:    string(p.Status),
			UserID:    p.UserID,
			CreatedAt: p.CreatedAt.Format(time.RFC3339),
		}
		if p.ThumbnailPath.Valid {
			item.ThumbnailURL = s.baseURL + p.ThumbnailPath.String
		}
		if u, ok := users[p.UserID]; ok {
			item.Username = u.Username
		}
		list[i] = item
	}

	return &ListReviewsResponse{
		List: list,
		Pagination: Pagination{
			Page:       result.Page,
			PageSize:   result.PageSize,
			Total:      result.Total,
			TotalPages: result.TotalPages,
		},
	}, nil
}

// ReviewRequest represents request for reviewing a photo
type ReviewRequest struct {
	Action string `json:"action" binding:"required,oneof=approve reject"`
	Reason string `json:"reason"`
}

// ReviewPhoto performs a manual review on a photo
func (s *Service) ReviewPhoto(ctx context.Context, photoID, reviewerID int64, req *ReviewRequest) error {
	// Check if photo exists
	exists, err := s.photoRepo.Exists(ctx, photoID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrPhotoNotFound
	}

	return s.photoRepo.ReviewPhoto(ctx, photoID, reviewerID, req.Action, req.Reason)
}

// ============================================
// Admin Delete Photo
// ============================================

// DeletePhotoRequest represents request for admin deleting a photo
type DeletePhotoRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// AdminDeletePhoto deletes a photo with reason
func (s *Service) AdminDeletePhoto(ctx context.Context, photoID, adminID int64, req *DeletePhotoRequest) error {
	err := s.photoRepo.AdminDeletePhoto(ctx, photoID, adminID, req.Reason)
	if errors.Is(err, postgresql.ErrNotFound) {
		return ErrPhotoNotFound
	}
	return err
}

// ============================================
// Ticket Management Methods
// ============================================

// AdminListTicketsRequest represents request for admin listing tickets
type AdminListTicketsRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Status   string `form:"status"`
	Type     string `form:"type"`
	UserID   int64  `form:"user_id"`
}

// TicketListItem represents a ticket in admin list
type TicketListItem struct {
	ID        int64   `json:"id"`
	Type      string  `json:"type"`
	Title     string  `json:"title"`
	Status    string  `json:"status"`
	UserID    int64   `json:"user_id"`
	Username  string  `json:"username"`
	PhotoID   *int64  `json:"photo_id,omitempty"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// AdminListTicketsResponse represents response for admin listing tickets
type AdminListTicketsResponse struct {
	List       []TicketListItem `json:"list"`
	Pagination Pagination       `json:"pagination"`
}

// AdminListTickets retrieves all tickets for admin
func (s *Service) AdminListTickets(ctx context.Context, req *AdminListTicketsRequest) (*AdminListTicketsResponse, error) {
	result, err := s.ticketRepo.AdminList(ctx, ticket.AdminListParams{
		Page:     req.Page,
		PageSize: req.PageSize,
		Status:   req.Status,
		Type:     req.Type,
		UserID:   req.UserID,
	})
	if err != nil {
		return nil, err
	}

	// Get user IDs
	userIDs := make([]int64, 0, len(result.Tickets))
	userIDMap := make(map[int64]bool)
	for _, t := range result.Tickets {
		if !userIDMap[t.UserID] {
			userIDs = append(userIDs, t.UserID)
			userIDMap[t.UserID] = true
		}
	}

	// Get users
	usersMap, err := s.ticketRepo.GetUserBriefMap(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	list := make([]TicketListItem, len(result.Tickets))
	for i, t := range result.Tickets {
		item := TicketListItem{
			ID:        t.ID,
			Type:      string(t.Type),
			Title:     t.Title,
			Status:    string(t.Status),
			UserID:    t.UserID,
			CreatedAt: t.CreatedAt.Format(time.RFC3339),
			UpdatedAt: t.UpdatedAt.Format(time.RFC3339),
		}
		if t.PhotoID.Valid {
			item.PhotoID = &t.PhotoID.Int64
		}
		if u, ok := usersMap[t.UserID]; ok {
			item.Username = u.Username
		}
		list[i] = item
	}

	return &AdminListTicketsResponse{
		List: list,
		Pagination: Pagination{
			Page:       result.Page,
			PageSize:   result.PageSize,
			Total:      result.Total,
			TotalPages: result.TotalPages,
		},
	}, nil
}

// ProcessTicketRequest represents request for processing a ticket
type ProcessTicketRequest struct {
	Status  string `json:"status" binding:"required,oneof=processing resolved closed"`
	Reply   string `json:"reply"`
}

// ProcessTicketResponse represents response for processing a ticket
type ProcessTicketResponse struct {
	ReplyID *int64 `json:"reply_id,omitempty"`
	Status  string `json:"status"`
}

// ProcessTicket processes a ticket (update status and optionally reply)
func (s *Service) ProcessTicket(ctx context.Context, ticketID, adminID int64, req *ProcessTicketRequest) (*ProcessTicketResponse, error) {
	// Check if ticket exists
	exists, err := s.ticketRepo.Exists(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrTicketNotFound
	}

	newStatus := model.TicketStatus(req.Status)
	resp := &ProcessTicketResponse{
		Status: req.Status,
	}

	if req.Reply != "" {
		// Reply and update status
		replyID, err := s.ticketRepo.AdminReply(ctx, ticketID, adminID, req.Reply, &newStatus)
		if err != nil {
			return nil, err
		}
		resp.ReplyID = &replyID
	} else {
		// Just update status
		err := s.ticketRepo.AdminUpdateStatus(ctx, ticketID, newStatus)
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}

// ============================================
// Featured Photos Methods
// ============================================

// AddFeaturedRequest represents request for adding a featured photo
type AddFeaturedRequest struct {
	PhotoID   int64  `json:"photo_id" binding:"required"`
	Reason    string `json:"reason"`
	SortOrder int    `json:"sort_order"`
}

// AddFeaturedResponse represents response for adding a featured photo
type AddFeaturedResponse struct {
	PhotoID int64 `json:"photo_id"`
	Message string `json:"message"`
}

// AddFeatured adds a photo to featured list
func (s *Service) AddFeatured(ctx context.Context, adminID int64, req *AddFeaturedRequest) (*AddFeaturedResponse, error) {
	// Check if photo exists
	exists, err := s.photoRepo.Exists(ctx, req.PhotoID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrPhotoNotFound
	}

	// Check if already featured
	isFeatured, err := s.photoRepo.IsFeaturedPhoto(ctx, req.PhotoID)
	if err != nil {
		return nil, err
	}
	if isFeatured {
		return nil, ErrAlreadyFeatured
	}

	err = s.photoRepo.AddFeatured(ctx, req.PhotoID, adminID, req.Reason, req.SortOrder)
	if err != nil {
		if errors.Is(err, postgresql.ErrDuplicateKey) {
			return nil, ErrAlreadyFeatured
		}
		return nil, err
	}

	return &AddFeaturedResponse{
		PhotoID: req.PhotoID,
		Message: "Photo added to featured list",
	}, nil
}

// RemoveFeatured removes a photo from featured list
func (s *Service) RemoveFeatured(ctx context.Context, photoID int64) error {
	err := s.photoRepo.RemoveFeaturedByPhotoID(ctx, photoID)
	if errors.Is(err, postgresql.ErrNotFound) {
		return ErrNotFeatured
	}
	return err
}

// ============================================
// Announcement Methods
// ============================================

// ListAnnouncementsRequest represents request for listing announcements
type ListAnnouncementsRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Status   string `form:"status"` // draft, published, all
}

// AnnouncementListItem represents an announcement in list
type AnnouncementListItem struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Summary     *string `json:"summary,omitempty"`
	Status      string  `json:"status"`
	IsPinned    bool    `json:"is_pinned"`
	AuthorID    int64   `json:"author_id"`
	PublishedAt *string `json:"published_at,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// ListAnnouncementsResponse represents response for listing announcements
type ListAnnouncementsResponse struct {
	List       []AnnouncementListItem `json:"list"`
	Pagination Pagination             `json:"pagination"`
}

// ListAnnouncements retrieves announcements for admin
func (s *Service) ListAnnouncements(ctx context.Context, req *ListAnnouncementsRequest) (*ListAnnouncementsResponse, error) {
	result, err := s.photoRepo.ListAnnouncements(ctx, photo.AnnouncementListParams{
		Page:     req.Page,
		PageSize: req.PageSize,
		Status:   req.Status,
	})
	if err != nil {
		return nil, err
	}

	list := make([]AnnouncementListItem, len(result.Announcements))
	for i, a := range result.Announcements {
		item := AnnouncementListItem{
			ID:        a.ID,
			Title:     a.Title,
			Status:    a.Status,
			IsPinned:  a.IsPinned,
			AuthorID:  a.AuthorID,
			CreatedAt: a.CreatedAt.Format(time.RFC3339),
			UpdatedAt: a.UpdatedAt.Format(time.RFC3339),
		}
		if a.Summary.Valid {
			item.Summary = &a.Summary.String
		}
		if a.PublishedAt.Valid {
			publishedAt := a.PublishedAt.Time.Format(time.RFC3339)
			item.PublishedAt = &publishedAt
		}
		list[i] = item
	}

	return &ListAnnouncementsResponse{
		List: list,
		Pagination: Pagination{
			Page:       result.Page,
			PageSize:   result.PageSize,
			Total:      result.Total,
			TotalPages: result.TotalPages,
		},
	}, nil
}

// CreateAnnouncementRequest represents request for creating an announcement
type CreateAnnouncementRequest struct {
	Title    string `json:"title" binding:"required,max=200"`
	Summary  string `json:"summary" binding:"max=500"`
	Content  string `json:"content" binding:"required"`
	Status   string `json:"status" binding:"required,oneof=draft published"`
	IsPinned bool   `json:"is_pinned"`
}

// AnnouncementResponse represents an announcement detail
type AnnouncementResponse struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Summary     *string `json:"summary,omitempty"`
	Content     string  `json:"content"`
	Status      string  `json:"status"`
	IsPinned    bool    `json:"is_pinned"`
	AuthorID    int64   `json:"author_id"`
	PublishedAt *string `json:"published_at,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// CreateAnnouncement creates a new announcement
func (s *Service) CreateAnnouncement(ctx context.Context, authorID int64, req *CreateAnnouncementRequest) (*AnnouncementResponse, error) {
	id, err := s.photoRepo.CreateAnnouncement(ctx, authorID, req.Title, req.Summary, req.Content, req.Status, req.IsPinned)
	if err != nil {
		return nil, err
	}

	// Get created announcement
	ann, err := s.photoRepo.GetAnnouncementByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.announcementToResponse(ann), nil
}

// UpdateAnnouncementRequest represents request for updating an announcement
type UpdateAnnouncementRequest struct {
	Title    string `json:"title" binding:"required,max=200"`
	Summary  string `json:"summary" binding:"max=500"`
	Content  string `json:"content" binding:"required"`
	Status   string `json:"status" binding:"required,oneof=draft published"`
	IsPinned bool   `json:"is_pinned"`
}

// UpdateAnnouncement updates an announcement
func (s *Service) UpdateAnnouncement(ctx context.Context, id int64, req *UpdateAnnouncementRequest) (*AnnouncementResponse, error) {
	err := s.photoRepo.UpdateAnnouncement(ctx, id, req.Title, req.Summary, req.Content, req.Status, req.IsPinned)
	if errors.Is(err, postgresql.ErrNotFound) {
		return nil, ErrAnnouncementNotFound
	}
	if err != nil {
		return nil, err
	}

	// Get updated announcement
	ann, err := s.photoRepo.GetAnnouncementByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return s.announcementToResponse(ann), nil
}

// DeleteAnnouncement deletes an announcement
func (s *Service) DeleteAnnouncement(ctx context.Context, id int64) error {
	err := s.photoRepo.DeleteAnnouncement(ctx, id)
	if errors.Is(err, postgresql.ErrNotFound) {
		return ErrAnnouncementNotFound
	}
	return err
}

// GetAnnouncement retrieves an announcement by ID
func (s *Service) GetAnnouncement(ctx context.Context, id int64) (*AnnouncementResponse, error) {
	ann, err := s.photoRepo.GetAnnouncementByID(ctx, id)
	if errors.Is(err, postgresql.ErrNotFound) {
		return nil, ErrAnnouncementNotFound
	}
	if err != nil {
		return nil, err
	}
	return s.announcementToResponse(ann), nil
}

func (s *Service) announcementToResponse(a *photo.Announcement) *AnnouncementResponse {
	resp := &AnnouncementResponse{
		ID:        a.ID,
		Title:     a.Title,
		Content:   a.Content,
		Status:    a.Status,
		IsPinned:  a.IsPinned,
		AuthorID:  a.AuthorID,
		CreatedAt: a.CreatedAt.Format(time.RFC3339),
		UpdatedAt: a.UpdatedAt.Format(time.RFC3339),
	}
	if a.Summary.Valid {
		resp.Summary = &a.Summary.String
	}
	if a.PublishedAt.Valid {
		publishedAt := a.PublishedAt.Time.Format(time.RFC3339)
		resp.PublishedAt = &publishedAt
	}
	return resp
}
