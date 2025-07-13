package services

import (
	"testing"

	"github.com/rogerwesterbo/svennescamping-backend/pkg/entities"
	"github.com/spf13/viper"
)

func TestRoleService_GetUserRole(t *testing.T) {
	// Set up test environment variables
	viper.Set("ADMIN_EMAILS", "admin@test.com,admin2@test.com")
	viper.Set("USER_EMAILS", "user@test.com,user2@test.com")

	// Create a new role service
	rs := NewRoleService()

	tests := []struct {
		name     string
		user     *entities.User
		expected entities.Role
	}{
		{
			name: "Admin user from email list",
			user: &entities.User{
				Email:    "admin@test.com",
				Verified: true,
			},
			expected: entities.RoleAdmin,
		},
		{
			name: "Regular user from email list",
			user: &entities.User{
				Email:    "user@test.com",
				Verified: true,
			},
			expected: entities.RoleUser,
		},
		{
			name: "User with svennescamping.no domain",
			user: &entities.User{
				Email:    "someone@svennescamping.no",
				Verified: true,
			},
			expected: entities.RoleNoAccess,
		},
		{
			name: "Verified user from unknown domain",
			user: &entities.User{
				Email:    "unknown@example.com",
				Verified: true,
			},
			expected: entities.RoleNoAccess,
		},
		{
			name: "Unverified user",
			user: &entities.User{
				Email:    "unverified@example.com",
				Verified: false,
			},
			expected: entities.RoleNoAccess,
		},
		{
			name: "User with admin group",
			user: &entities.User{
				Email:    "groupadmin@example.com",
				Verified: true,
				Groups:   []string{"admin"},
			},
			expected: entities.RoleAdmin,
		},
		{
			name: "User with user group",
			user: &entities.User{
				Email:    "groupuser@example.com",
				Verified: true,
				Groups:   []string{"user"},
			},
			expected: entities.RoleUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rs.GetUserRole(tt.user)
			if result != tt.expected {
				t.Errorf("GetUserRole() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRoleService_EmailLists(t *testing.T) {
	// Set up test environment variables
	viper.Set("ADMIN_EMAILS", "admin1@test.com,admin2@test.com,admin3@test.com")
	viper.Set("USER_EMAILS", "user1@test.com,user2@test.com")

	// Create a new role service
	rs := NewRoleService()

	// Test admin emails
	adminEmails := rs.GetAdminEmails()
	expectedAdmins := []string{"admin1@test.com", "admin2@test.com", "admin3@test.com"}
	if len(adminEmails) != len(expectedAdmins) {
		t.Errorf("Expected %d admin emails, got %d", len(expectedAdmins), len(adminEmails))
	}

	// Test user emails
	userEmails := rs.GetUserEmails()
	expectedUsers := []string{"user1@test.com", "user2@test.com"}
	if len(userEmails) != len(expectedUsers) {
		t.Errorf("Expected %d user emails, got %d", len(expectedUsers), len(userEmails))
	}
}
