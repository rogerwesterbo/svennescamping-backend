package entities

// User represents a user extracted from Google OAuth access token
type User struct {
	ID       string   `json:"id"`       // Google user ID
	Email    string   `json:"email"`    // User's email address
	Name     string   `json:"name"`     // User's full name
	Picture  string   `json:"picture"`  // User's profile picture URL
	Groups   []string `json:"groups"`   // Groups the user belongs to (from token claims)
	Verified bool     `json:"verified"` // Whether the email is verified
	Role     Role     `json:"role"`     // User's role in the system
}

// GoogleTokenClaims represents the claims in a Google OAuth access token
type GoogleTokenClaims struct {
	ID            string   `json:"sub"`            // Subject (user ID)
	Email         string   `json:"email"`          // Email address
	Name          string   `json:"name"`           // Full name
	Picture       string   `json:"picture"`        // Profile picture URL
	EmailVerified bool     `json:"email_verified"` // Email verification status
	Groups        []string `json:"groups"`         // Custom groups claim
	Audience      string   `json:"aud"`            // Audience
	Issuer        string   `json:"iss"`            // Issuer
	ExpiresAt     int64    `json:"exp"`            // Expiration time
	IssuedAt      int64    `json:"iat"`            // Issued at time
}

// ToUser converts GoogleTokenClaims to User entity
func (claims *GoogleTokenClaims) ToUser() *User {
	return &User{
		ID:       claims.ID,
		Email:    claims.Email,
		Name:     claims.Name,
		Picture:  claims.Picture,
		Groups:   claims.Groups,
		Verified: claims.EmailVerified,
		Role:     RoleUser, // Default role, can be overridden based on business logic
	}
}

// HasPermission checks if the user has a specific permission
func (u *User) HasPermission(permission Permission) bool {
	return u.Role.HasPermission(permission)
}

// IsAdmin checks if the user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// IsUser checks if the user has user role
func (u *User) IsUser() bool {
	return u.Role == RoleUser
}

// HasAccess checks if the user has any access (not no_access role)
func (u *User) HasAccess() bool {
	return u.Role != RoleNoAccess
}
