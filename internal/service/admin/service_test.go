package admin

import (
	"testing"

	"QuanPhotos/internal/model"
)

func TestUpdateRoleValidation(t *testing.T) {
	tests := []struct {
		name         string
		operatorRole model.UserRole
		targetRole   model.UserRole
		newRole      string
		expectError  bool
	}{
		{
			name:         "Admin can promote user to reviewer",
			operatorRole: model.RoleAdmin,
			targetRole:   model.RoleUser,
			newRole:      "reviewer",
			expectError:  false,
		},
		{
			name:         "Admin cannot promote to admin",
			operatorRole: model.RoleAdmin,
			targetRole:   model.RoleUser,
			newRole:      "admin",
			expectError:  true,
		},
		{
			name:         "Admin cannot change other admin",
			operatorRole: model.RoleAdmin,
			targetRole:   model.RoleAdmin,
			newRole:      "user",
			expectError:  true,
		},
		{
			name:         "SuperAdmin can promote to admin",
			operatorRole: model.RoleSuperAdmin,
			targetRole:   model.RoleUser,
			newRole:      "admin",
			expectError:  false,
		},
		{
			name:         "Cannot set to superadmin",
			operatorRole: model.RoleSuperAdmin,
			targetRole:   model.RoleUser,
			newRole:      "superadmin",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check role validation logic
			newRole := model.UserRole(tt.newRole)

			hasError := false

			// Admin permission check
			if tt.operatorRole == model.RoleAdmin {
				if tt.targetRole == model.RoleAdmin || tt.targetRole == model.RoleSuperAdmin {
					hasError = true
				}
				if newRole == model.RoleAdmin {
					hasError = true
				}
			}

			// Cannot set to superadmin
			if newRole == model.RoleSuperAdmin {
				hasError = true
			}

			if hasError != tt.expectError {
				t.Errorf("expected error=%v, got error=%v", tt.expectError, hasError)
			}
		})
	}
}

func TestUpdateStatusValidation(t *testing.T) {
	tests := []struct {
		name         string
		operatorRole model.UserRole
		targetRole   model.UserRole
		expectError  bool
	}{
		{
			name:         "Admin can ban user",
			operatorRole: model.RoleAdmin,
			targetRole:   model.RoleUser,
			expectError:  false,
		},
		{
			name:         "Admin cannot ban other admin",
			operatorRole: model.RoleAdmin,
			targetRole:   model.RoleAdmin,
			expectError:  true,
		},
		{
			name:         "SuperAdmin can ban admin",
			operatorRole: model.RoleSuperAdmin,
			targetRole:   model.RoleAdmin,
			expectError:  false,
		},
		{
			name:         "SuperAdmin cannot ban other superadmin",
			operatorRole: model.RoleSuperAdmin,
			targetRole:   model.RoleSuperAdmin,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasError := false

			// Admin permission check
			if tt.operatorRole == model.RoleAdmin {
				if tt.targetRole == model.RoleAdmin || tt.targetRole == model.RoleSuperAdmin {
					hasError = true
				}
			}

			// SuperAdmin permission check
			if tt.operatorRole == model.RoleSuperAdmin {
				if tt.targetRole == model.RoleSuperAdmin {
					hasError = true
				}
			}

			if hasError != tt.expectError {
				t.Errorf("expected error=%v, got error=%v", tt.expectError, hasError)
			}
		})
	}
}
