package auth

import (
	"context"
	"errors"

	"QuanPhotos/internal/model"
	"QuanPhotos/internal/pkg/hash"
	"QuanPhotos/internal/repository/postgresql"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserBanned         = errors.New("user account is banned")
)

// Login authenticates a user and returns tokens
func (s *Service) Login(ctx context.Context, req *LoginRequest) (*model.User, *TokenPair, error) {
	// Get user by username
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, postgresql.ErrUserNotFound) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}

	// Check password
	if !hash.CheckPassword(req.Password, user.PasswordHash) {
		return nil, nil, ErrInvalidCredentials
	}

	// Check if user is banned
	if user.Status == model.StatusBanned {
		return nil, nil, ErrUserBanned
	}

	// Update last login time
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	// Generate tokens
	tokens, err := s.generateTokenPair(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

// LoginByEmail authenticates a user by email and returns tokens
func (s *Service) LoginByEmail(ctx context.Context, email, password string) (*model.User, *TokenPair, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, postgresql.ErrUserNotFound) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, err
	}

	// Check password
	if !hash.CheckPassword(password, user.PasswordHash) {
		return nil, nil, ErrInvalidCredentials
	}

	// Check if user is banned
	if user.Status == model.StatusBanned {
		return nil, nil, ErrUserBanned
	}

	// Update last login time
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID)

	// Generate tokens
	tokens, err := s.generateTokenPair(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}
