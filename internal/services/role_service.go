package services

import (
	"strings"

	"github.com/rogerwesterbo/svennescamping-backend/pkg/consts"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/entities"
	"github.com/spf13/viper"
)

// RoleService handles role assignment and management
type RoleService struct {
	// In-memory role mappings (in production, this would come from a database)
	userRoles   map[string]entities.Role
	adminEmails []string
	usersEmails []string
}

// NewRoleService creates a new role service
func NewRoleService() *RoleService {
	// Read admin emails from environment variable
	adminEmailsStr := viper.GetString(consts.ADMIN_EMAILS)
	var adminEmails []string
	if adminEmailsStr != "" {
		adminEmails = strings.Split(adminEmailsStr, ",")
		// Trim whitespace from each email
		for i, email := range adminEmails {
			adminEmails[i] = strings.TrimSpace(email)
		}
	}

	// Read user emails from environment variable
	userEmailsStr := viper.GetString(consts.USER_EMAILS)
	var userEmails []string
	if userEmailsStr != "" {
		userEmails = strings.Split(userEmailsStr, ",")
		// Trim whitespace from each email
		for i, email := range userEmails {
			userEmails[i] = strings.TrimSpace(email)
		}
	}

	// Debug logging to verify environment variables are loaded
	// logger.Info("Role service initialized",
	// 	zap.String("admin_emails_raw", adminEmailsStr),
	// 	zap.Strings("admin_emails", adminEmails),
	// 	zap.String("user_emails_raw", userEmailsStr),
	// 	zap.Strings("user_emails", userEmails),
	// )

	return &RoleService{
		userRoles:   make(map[string]entities.Role),
		adminEmails: adminEmails,
		usersEmails: userEmails,
	}
}

// GetUserRole determines the role for a user based on their email and other criteria
func (rs *RoleService) GetUserRole(user *entities.User) entities.Role {
	// Check if user is in the admin list
	if rs.isAdminEmail(user.Email) {
		return entities.RoleAdmin
	}

	// Check if user is in the user list
	if rs.isUserEmail(user.Email) {
		return entities.RoleUser
	}

	// Check if there's a specific role assignment for this user
	if role, exists := rs.userRoles[user.Email]; exists {
		return role
	}

	// Check if user belongs to specific groups that grant admin access
	for _, group := range user.Groups {
		if strings.ToLower(group) == "admin" || strings.ToLower(group) == "administrators" {
			return entities.RoleAdmin
		}
		if strings.ToLower(group) == "user" || strings.ToLower(group) == "users" {
			return entities.RoleUser
		}
		if strings.ToLower(group) == "no_access" || strings.ToLower(group) == "noaccess" {
			return entities.RoleNoAccess
		}
	}

	// No access for unverified emails or unknown domains
	return entities.RoleNoAccess
}

// SetUserRole manually sets a role for a specific user
func (rs *RoleService) SetUserRole(email string, role entities.Role) {
	if role.IsValid() {
		rs.userRoles[email] = role
	}
}

// RemoveUserRole removes a specific role assignment
func (rs *RoleService) RemoveUserRole(email string) {
	delete(rs.userRoles, email)
}

// isAdminEmail checks if an email is in the admin list
func (rs *RoleService) isAdminEmail(email string) bool {
	for _, adminEmail := range rs.adminEmails {
		if strings.EqualFold(email, adminEmail) {
			return true
		}
	}
	return false
}

// isUsersEmail checks if an email is in the admin list
func (rs *RoleService) isUserEmail(email string) bool {
	for _, userEmail := range rs.usersEmails {
		if strings.EqualFold(email, userEmail) {
			return true
		}
	}
	return false
}

// AddAdminEmail adds an email to the admin list
func (rs *RoleService) AddAdminEmail(email string) {
	rs.adminEmails = append(rs.adminEmails, email)
}

// AddUserEmail adds an email to the user list
func (rs *RoleService) AddUserEmail(email string) {
	rs.usersEmails = append(rs.usersEmails, email)
}

// GetAdminEmails returns the list of admin emails
func (rs *RoleService) GetAdminEmails() []string {
	return rs.adminEmails
}

// GetUserEmails returns the list of user emails
func (rs *RoleService) GetUserEmails() []string {
	return rs.usersEmails
}

// GetAllUserRoles returns all user role assignments
func (rs *RoleService) GetAllUserRoles() map[string]entities.Role {
	result := make(map[string]entities.Role)
	for email, role := range rs.userRoles {
		result[email] = role
	}
	return result
}
