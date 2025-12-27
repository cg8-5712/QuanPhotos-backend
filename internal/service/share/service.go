package share

import (
	"context"
	"errors"
	"fmt"

	"QuanPhotos/internal/repository/postgresql/share"
)

var (
	ErrPhotoNotFound = errors.New("photo not found")
)

// Service handles share business logic
type Service struct {
	shareRepo *share.ShareRepository
	baseURL   string
}

// New creates a new share service
func New(shareRepo *share.ShareRepository, baseURL string) *Service {
	return &Service{
		shareRepo: shareRepo,
		baseURL:   baseURL,
	}
}

// ShareRequest represents request for sharing a photo
type ShareRequest struct {
	PhotoID  int64  `json:"-"`
	Platform string `json:"platform" binding:"required,oneof=twitter facebook weibo wechat link"`
	Note     string `json:"note" binding:"max=500"`
}

// ShareResponse represents response for sharing a photo
type ShareResponse struct {
	ShareURL   string `json:"share_url"`
	ShareCount int    `json:"share_count"`
}

// Share creates a share record and returns share info
func (s *Service) Share(ctx context.Context, userID int64, req *ShareRequest) (*ShareResponse, error) {
	// Check if photo exists
	exists, err := s.shareRepo.PhotoExists(ctx, req.PhotoID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrPhotoNotFound
	}

	// Create share record
	_, err = s.shareRepo.Create(ctx, req.PhotoID, userID, req.Platform, req.Note)
	if err != nil {
		return nil, err
	}

	// Increment share count
	err = s.shareRepo.IncrementShareCount(ctx, req.PhotoID)
	if err != nil {
		// Log but don't fail
	}

	// Get updated share count
	shareCount, err := s.shareRepo.GetShareCount(ctx, req.PhotoID)
	if err != nil {
		shareCount = 0
	}

	// Generate share URL
	shareURL := s.generateShareURL(req.PhotoID)

	return &ShareResponse{
		ShareURL:   shareURL,
		ShareCount: shareCount,
	}, nil
}

// generateShareURL generates a shareable URL for a photo
func (s *Service) generateShareURL(photoID int64) string {
	return fmt.Sprintf("%s/photos/%d", s.baseURL, photoID)
}
