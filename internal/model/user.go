package model

import (
	"database/sql"
	"time"
)

// UserRole represents user roles
type UserRole string

const (
	RoleGuest      UserRole = "guest"
	RoleUser       UserRole = "user"
	RoleReviewer   UserRole = "reviewer"
	RoleAdmin      UserRole = "admin"
	RoleSuperAdmin UserRole = "superadmin"
)

// UserStatus represents user status
type UserStatus string

const (
	StatusActive UserStatus = "active"
	StatusBanned UserStatus = "banned"
)

// User represents a user in the system
type User struct {
	ID           int64          `db:"id" json:"id"`
	Username     string         `db:"username" json:"username"`
	Email        string         `db:"email" json:"email"`
	PasswordHash string         `db:"password_hash" json:"-"`
	Role         UserRole       `db:"role" json:"role"`
	Status       UserStatus     `db:"status" json:"status"`
	CanComment   bool           `db:"can_comment" json:"can_comment"`
	CanMessage   bool           `db:"can_message" json:"can_message"`
	CanUpload    bool           `db:"can_upload" json:"can_upload"`
	Avatar       sql.NullString `db:"avatar" json:"-"`
	Bio          sql.NullString `db:"bio" json:"-"`
	Location     sql.NullString `db:"location" json:"-"`
	LastLoginAt  sql.NullTime   `db:"last_login_at" json:"-"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at" json:"updated_at"`
}

// UserPublicInfo represents public user information
type UserPublicInfo struct {
	ID        int64    `json:"id"`
	Username  string   `json:"username"`
	Avatar    *string  `json:"avatar"`
	Bio       *string  `json:"bio"`
	Location  *string  `json:"location"`
	Role      UserRole `json:"role"`
	CreatedAt string   `json:"created_at"`
}

// UserProfile represents full user profile (for self)
type UserProfile struct {
	ID          int64      `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	Role        UserRole   `json:"role"`
	Status      UserStatus `json:"status"`
	CanComment  bool       `json:"can_comment"`
	CanMessage  bool       `json:"can_message"`
	CanUpload   bool       `json:"can_upload"`
	Avatar      *string    `json:"avatar"`
	Bio         *string    `json:"bio"`
	Location    *string    `json:"location"`
	LastLoginAt *string    `json:"last_login_at"`
	CreatedAt   string     `json:"created_at"`
	UpdatedAt   string     `json:"updated_at"`
}

// ToPublicInfo converts User to UserPublicInfo
func (u *User) ToPublicInfo() *UserPublicInfo {
	info := &UserPublicInfo{
		ID:        u.ID,
		Username:  u.Username,
		Role:      u.Role,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	}

	if u.Avatar.Valid {
		info.Avatar = &u.Avatar.String
	}
	if u.Bio.Valid {
		info.Bio = &u.Bio.String
	}
	if u.Location.Valid {
		info.Location = &u.Location.String
	}

	return info
}

// ToProfile converts User to UserProfile
func (u *User) ToProfile() *UserProfile {
	profile := &UserProfile{
		ID:         u.ID,
		Username:   u.Username,
		Email:      u.Email,
		Role:       u.Role,
		Status:     u.Status,
		CanComment: u.CanComment,
		CanMessage: u.CanMessage,
		CanUpload:  u.CanUpload,
		CreatedAt:  u.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  u.UpdatedAt.Format(time.RFC3339),
	}

	if u.Avatar.Valid {
		profile.Avatar = &u.Avatar.String
	}
	if u.Bio.Valid {
		profile.Bio = &u.Bio.String
	}
	if u.Location.Valid {
		profile.Location = &u.Location.String
	}
	if u.LastLoginAt.Valid {
		lastLogin := u.LastLoginAt.Time.Format(time.RFC3339)
		profile.LastLoginAt = &lastLogin
	}

	return profile
}

// IsActive checks if user is active
func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

// HasRole checks if user has the specified role
func (u *User) HasRole(role UserRole) bool {
	return u.Role == role
}

// IsAdmin checks if user is admin or superadmin
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin || u.Role == RoleSuperAdmin
}

// IsSuperAdmin checks if user is superadmin
func (u *User) IsSuperAdmin() bool {
	return u.Role == RoleSuperAdmin
}

// IsReviewer checks if user is reviewer or higher
func (u *User) IsReviewer() bool {
	return u.Role == RoleReviewer || u.Role == RoleAdmin || u.Role == RoleSuperAdmin
}

// CanPerformAction checks if user can perform the specified action
func (u *User) CanPerformAction(action string) bool {
	if !u.IsActive() {
		return false
	}

	switch action {
	case "comment":
		return u.CanComment
	case "message":
		return u.CanMessage
	case "upload":
		return u.CanUpload
	default:
		return true
	}
}

// RoleLevel returns the privilege level of a role (higher = more privileges)
func RoleLevel(role UserRole) int {
	switch role {
	case RoleGuest:
		return 0
	case RoleUser:
		return 1
	case RoleReviewer:
		return 2
	case RoleAdmin:
		return 3
	case RoleSuperAdmin:
		return 4
	default:
		return 0
	}
}

// HasMinRole checks if user has at least the specified role level
func (u *User) HasMinRole(minRole UserRole) bool {
	return RoleLevel(u.Role) >= RoleLevel(minRole)
}
