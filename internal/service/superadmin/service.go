package superadmin

import (
	"context"
	"errors"

	"QuanPhotos/internal/model"
	"QuanPhotos/internal/repository/postgresql"
	"QuanPhotos/internal/repository/postgresql/superadmin"
)

var (
	ErrInvalidPermission = errors.New("invalid permission")
	ErrNotAdmin          = errors.New("user is not an admin")
	ErrNotReviewer       = errors.New("user is not a reviewer")
	ErrCategoryNotFound  = errors.New("category not found")
	ErrUserNotFound      = errors.New("user not found")
	ErrSelfModify        = errors.New("cannot modify own permissions")
)

// Service handles superadmin business logic
type Service struct {
	repo *superadmin.SuperadminRepository
}

// New creates a new superadmin service
func New(repo *superadmin.SuperadminRepository) *Service {
	return &Service{repo: repo}
}

// AdminBrief is a brief representation of an admin user
type AdminBrief struct {
	ID          int64    `json:"id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Avatar      string   `json:"avatar"`
	Permissions []string `json:"permissions"`
}

// ReviewerBrief is a brief representation of a reviewer user
type ReviewerBrief struct {
	ID         int64           `json:"id"`
	Username   string          `json:"username"`
	Email      string          `json:"email"`
	Avatar     string          `json:"avatar"`
	Categories []CategoryBrief `json:"categories"`
}

// CategoryBrief is a brief representation of a category
type CategoryBrief struct {
	ID     int32  `json:"id"`
	Name   string `json:"name"`
	NameEN string `json:"name_en"`
}

// UserRestrictions represents user restriction settings
type UserRestrictions struct {
	UserID     int64  `json:"user_id"`
	Username   string `json:"username"`
	CanComment bool   `json:"can_comment"`
	CanMessage bool   `json:"can_message"`
	CanUpload  bool   `json:"can_upload"`
}

// ListAdminsResult is the result of listing admins
type ListAdminsResult struct {
	Admins []*AdminBrief `json:"admins"`
	Total  int           `json:"total"`
}

// ListReviewersResult is the result of listing reviewers
type ListReviewersResult struct {
	Reviewers []*ReviewerBrief `json:"reviewers"`
	Total     int              `json:"total"`
}

// ListAdmins retrieves all admin users with their permissions
func (s *Service) ListAdmins(ctx context.Context) (*ListAdminsResult, error) {
	admins, err := s.repo.ListAdmins(ctx)
	if err != nil {
		return nil, err
	}

	result := &ListAdminsResult{
		Admins: make([]*AdminBrief, len(admins)),
		Total:  len(admins),
	}

	for i, admin := range admins {
		result.Admins[i] = s.toAdminBrief(admin.User, admin.Permissions)
	}

	return result, nil
}

// GetAdminPermissions retrieves permissions for a specific admin
func (s *Service) GetAdminPermissions(ctx context.Context, adminID int64) ([]string, error) {
	// Verify user is an admin
	_, err := s.repo.GetAdminByID(ctx, adminID)
	if err != nil {
		if errors.Is(err, postgresql.ErrNotFound) {
			return nil, ErrNotAdmin
		}
		return nil, err
	}

	return s.repo.GetAdminPermissions(ctx, adminID)
}

// GrantPermissions grants permissions to an admin
func (s *Service) GrantPermissions(ctx context.Context, adminID int64, permissions []string, grantedBy int64) error {
	// Cannot modify own permissions
	if adminID == grantedBy {
		return ErrSelfModify
	}

	// Verify user is an admin
	_, err := s.repo.GetAdminByID(ctx, adminID)
	if err != nil {
		if errors.Is(err, postgresql.ErrNotFound) {
			return ErrNotAdmin
		}
		return err
	}

	// Validate permissions
	for _, perm := range permissions {
		if !isValidPermission(perm) {
			return ErrInvalidPermission
		}
	}

	// Grant each permission
	for _, perm := range permissions {
		err := s.repo.GrantPermission(ctx, adminID, perm, grantedBy)
		if err != nil {
			return err
		}
	}

	return nil
}

// RevokePermissions revokes permissions from an admin
func (s *Service) RevokePermissions(ctx context.Context, adminID int64, permissions []string, revokedBy int64) error {
	// Cannot modify own permissions
	if adminID == revokedBy {
		return ErrSelfModify
	}

	// Verify user is an admin
	_, err := s.repo.GetAdminByID(ctx, adminID)
	if err != nil {
		if errors.Is(err, postgresql.ErrNotFound) {
			return ErrNotAdmin
		}
		return err
	}

	// Revoke each permission
	for _, perm := range permissions {
		err := s.repo.RevokePermission(ctx, adminID, perm)
		if err != nil && !errors.Is(err, postgresql.ErrNotFound) {
			return err
		}
	}

	return nil
}

// ListReviewers retrieves all reviewer users with their assigned categories
func (s *Service) ListReviewers(ctx context.Context) (*ListReviewersResult, error) {
	reviewers, err := s.repo.ListReviewers(ctx)
	if err != nil {
		return nil, err
	}

	result := &ListReviewersResult{
		Reviewers: make([]*ReviewerBrief, len(reviewers)),
		Total:     len(reviewers),
	}

	for i, reviewer := range reviewers {
		result.Reviewers[i] = s.toReviewerBrief(reviewer.User, reviewer.Categories)
	}

	return result, nil
}

// GetReviewerCategories retrieves categories assigned to a specific reviewer
func (s *Service) GetReviewerCategories(ctx context.Context, reviewerID int64) ([]int, error) {
	// Verify user is a reviewer
	_, err := s.repo.GetReviewerByID(ctx, reviewerID)
	if err != nil {
		if errors.Is(err, postgresql.ErrNotFound) {
			return nil, ErrNotReviewer
		}
		return nil, err
	}

	return s.repo.GetReviewerCategories(ctx, reviewerID)
}

// AssignCategories assigns categories to a reviewer
func (s *Service) AssignCategories(ctx context.Context, reviewerID int64, categoryIDs []int) error {
	// Verify user is a reviewer
	_, err := s.repo.GetReviewerByID(ctx, reviewerID)
	if err != nil {
		if errors.Is(err, postgresql.ErrNotFound) {
			return ErrNotReviewer
		}
		return err
	}

	// Verify all categories exist
	for _, catID := range categoryIDs {
		exists, err := s.repo.CategoryExists(ctx, catID)
		if err != nil {
			return err
		}
		if !exists {
			return ErrCategoryNotFound
		}
	}

	// Assign each category
	for _, catID := range categoryIDs {
		err := s.repo.AssignCategory(ctx, reviewerID, catID)
		if err != nil {
			return err
		}
	}

	return nil
}

// RevokeCategories revokes categories from a reviewer
func (s *Service) RevokeCategories(ctx context.Context, reviewerID int64, categoryIDs []int) error {
	// Verify user is a reviewer
	_, err := s.repo.GetReviewerByID(ctx, reviewerID)
	if err != nil {
		if errors.Is(err, postgresql.ErrNotFound) {
			return ErrNotReviewer
		}
		return err
	}

	// Revoke each category
	for _, catID := range categoryIDs {
		err := s.repo.RevokeCategory(ctx, reviewerID, catID)
		if err != nil && !errors.Is(err, postgresql.ErrNotFound) {
			return err
		}
	}

	return nil
}

// GetUserRestrictions retrieves restriction settings for a user
func (s *Service) GetUserRestrictions(ctx context.Context, userID int64) (*UserRestrictions, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, postgresql.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	canComment, canMessage, canUpload, err := s.repo.GetUserRestrictions(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &UserRestrictions{
		UserID:     user.ID,
		Username:   user.Username,
		CanComment: canComment,
		CanMessage: canMessage,
		CanUpload:  canUpload,
	}, nil
}

// UpdateUserRestrictions updates restriction settings for a user
func (s *Service) UpdateUserRestrictions(ctx context.Context, userID int64, canComment, canMessage, canUpload *bool) (*UserRestrictions, error) {
	// Verify user exists
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, postgresql.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	// Don't allow modifying superadmin restrictions
	if user.Role == model.RoleSuperAdmin {
		return nil, errors.New("cannot modify superadmin restrictions")
	}

	err = s.repo.UpdateUserRestrictions(ctx, userID, canComment, canMessage, canUpload)
	if err != nil {
		return nil, err
	}

	// Get updated restrictions
	return s.GetUserRestrictions(ctx, userID)
}

// GetAllPermissions returns all available permission names
func (s *Service) GetAllPermissions() []string {
	return superadmin.AllPermissions
}

// Helper functions

func (s *Service) toAdminBrief(user *model.User, permissions []string) *AdminBrief {
	return &AdminBrief{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		Avatar:      user.Avatar.String,
		Permissions: permissions,
	}
}

func (s *Service) toReviewerBrief(user *model.User, categories []*model.Category) *ReviewerBrief {
	cats := make([]CategoryBrief, len(categories))
	for i, cat := range categories {
		cats[i] = CategoryBrief{
			ID:     cat.ID,
			Name:   cat.Name,
			NameEN: cat.NameEN,
		}
	}

	return &ReviewerBrief{
		ID:         user.ID,
		Username:   user.Username,
		Email:      user.Email,
		Avatar:     user.Avatar.String,
		Categories: cats,
	}
}

func isValidPermission(perm string) bool {
	for _, p := range superadmin.AllPermissions {
		if p == perm {
			return true
		}
	}
	return false
}
