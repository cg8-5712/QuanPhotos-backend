package admin

import (
	"context"
	"errors"

	"QuanPhotos/internal/model"
	"QuanPhotos/internal/repository/postgresql"
	"QuanPhotos/internal/repository/postgresql/user"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrCannotChangeSelf = errors.New("cannot change own role or status")
	ErrInvalidRole      = errors.New("invalid role")
	ErrInsufficientPerm = errors.New("insufficient permissions")
)

// Service handles admin business logic
type Service struct {
	userRepo *user.UserRepository
}

// New creates a new admin service
func New(userRepo *user.UserRepository) *Service {
	return &Service{
		userRepo: userRepo,
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
