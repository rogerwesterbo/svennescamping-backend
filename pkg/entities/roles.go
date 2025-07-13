package entities

// Role represents a user role in the system
type Role string

const (
	// RoleAdmin has full access to all resources
	RoleAdmin Role = "admin"
	// RoleUser has access to user-level resources
	RoleUser Role = "user"
	// RoleNoAccess has no access to protected resources
	RoleNoAccess Role = "no_access"
)

// Permission represents a specific permission
type Permission string

const (
	// User permissions
	PermissionReadOwnProfile   Permission = "read:own_profile"
	PermissionUpdateOwnProfile Permission = "update:own_profile"

	// Transaction permissions
	PermissionReadTransactions   Permission = "read:transactions"
	PermissionCreateTransactions Permission = "create:transactions"
	PermissionUpdateTransactions Permission = "update:transactions"
	PermissionDeleteTransactions Permission = "delete:transactions"

	// Admin permissions
	PermissionReadAllUsers Permission = "read:all_users"
	PermissionManageUsers  Permission = "manage:users"
	PermissionSystemAdmin  Permission = "system:admin"
)

// RolePermissions maps roles to their permissions
var RolePermissions = map[Role][]Permission{
	RoleAdmin: {
		// Admin has all permissions
		PermissionReadOwnProfile,
		PermissionUpdateOwnProfile,
		PermissionReadTransactions,
		PermissionCreateTransactions,
		PermissionUpdateTransactions,
		PermissionDeleteTransactions,
		PermissionReadAllUsers,
		PermissionManageUsers,
		PermissionSystemAdmin,
	},
	RoleUser: {
		// User has limited permissions
		PermissionReadOwnProfile,
		PermissionUpdateOwnProfile,
		PermissionReadTransactions,
	},
	RoleNoAccess: {
		// No access has no permissions
	},
}

// HasPermission checks if a role has a specific permission
func (r Role) HasPermission(permission Permission) bool {
	permissions, exists := RolePermissions[r]
	if !exists {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// IsValid checks if the role is a valid role
func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleUser, RoleNoAccess:
		return true
	default:
		return false
	}
}

// String returns the string representation of the role
func (r Role) String() string {
	return string(r)
}
