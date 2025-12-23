package user

import (
	"context"
	"errors"

	"QuanPhotos/internal/model"
	"QuanPhotos/internal/pkg/hash"
	"QuanPhotos/internal/repository/postgresql"
	"QuanPhotos/internal/repository/postgresql/user"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidPassword = errors.New("invalid current password")
)

// Service handles user business logic
type Service struct {
	userRepo *user.UserRepository
}

// New creates a new user service
func New(userRepo *user.UserRepository) *Service {
	return &Service{
		userRepo: userRepo,
	}
}

// GetByID retrieves a user by ID
func (s *Service) GetByID(ctx context.Context, id int64) (*model.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, postgresql.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// GetProfile retrieves user profile
func (s *Service) GetProfile(ctx context.Context, userID int64) (*model.UserProfile, error) {
	user, err := s.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user.ToProfile(), nil
}

// GetPublicInfo retrieves public user information
func (s *Service) GetPublicInfo(ctx context.Context, userID int64) (*model.UserPublicInfo, error) {
	user, err := s.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user.ToPublicInfo(), nil
}

// UpdateProfileRequest represents profile update request
type UpdateProfileRequest struct {
	Avatar   *string `json:"avatar"`
	Bio      *string `json:"bio"`
	Location *string `json:"location"`
}

// UpdateProfile updates user's profile
func (s *Service) UpdateProfile(ctx context.Context, userID int64, req *UpdateProfileRequest) (*model.UserProfile, error) {
	if err := s.userRepo.UpdateProfile(ctx, userID, req.Avatar, req.Bio, req.Location); err != nil {
		if errors.Is(err, postgresql.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return s.GetProfile(ctx, userID)
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8,max=72"`
}

// ChangePassword changes user's password
func (s *Service) ChangePassword(ctx context.Context, userID int64, req *ChangePasswordRequest) error {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, postgresql.ErrUserNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	// Verify current password
	if !hash.CheckPassword(req.CurrentPassword, user.PasswordHash) {
		return ErrInvalidPassword
	}

	// Hash new password
	newHash, err := hash.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	// Update password
	return s.userRepo.UpdatePassword(ctx, userID, newHash)
}
