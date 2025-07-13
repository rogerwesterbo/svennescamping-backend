package middlewares

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rogerwesterbo/svennescamping-backend/internal/services"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/entities"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/helpers/httphelpers"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/logger"
	"go.uber.org/zap"
)

// UserContextKey is the key used to store user information in request context
type UserContextKey string

const UserKey UserContextKey = "user"

// Global role service instance (initialized after settings)
var roleService *services.RoleService

// InitializeRoleService initializes the role service after settings are loaded
func InitializeRoleService() {
	roleService = services.NewRoleService()
}

// getRoleService returns the role service instance
func getRoleService() *services.RoleService {
	if roleService == nil {
		// Fallback: create a new instance if not initialized (shouldn't happen in normal flow)
		roleService = services.NewRoleService()
	}
	return roleService
}

// AuthMiddleware verifies Google OAuth access tokens
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logger.Warn("Missing Authorization header",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
			)
			httphelpers.RespondWithJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Missing Authorization header",
			})
			return
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			logger.Warn("Invalid Authorization header format",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
			)
			httphelpers.RespondWithJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Invalid Authorization header format",
			})
			return
		}

		// Extract the token
		accessToken := strings.TrimPrefix(authHeader, "Bearer ")
		if accessToken == "" {
			logger.Warn("Empty access token",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
			)
			httphelpers.RespondWithJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Empty access token",
			})
			return
		}

		// Verify the token with Google
		user, err := verifyGoogleAccessToken(accessToken)
		if err != nil {
			logger.Warn("Token verification failed",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
				zap.Error(err),
			)
			httphelpers.RespondWithJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Invalid or expired access token",
			})
			return
		}

		// Add user to request context
		ctx := context.WithValue(r.Context(), UserKey, user)
		r = r.WithContext(ctx)

		logger.Info("User authenticated successfully",
			zap.String("userID", user.ID),
			zap.String("email", user.Email),
			zap.String("path", r.URL.Path),
		)

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// verifyGoogleAccessToken verifies the access token with Google's tokeninfo endpoint
func verifyGoogleAccessToken(accessToken string) (*entities.User, error) {
	// Use Google's tokeninfo endpoint to verify the access token
	url := fmt.Sprintf("https://www.googleapis.com/oauth2/v1/tokeninfo?access_token=%s", accessToken)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token verification failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read verification response: %w", err)
	}

	// Parse the tokeninfo response
	var tokenInfo map[string]interface{}
	if err := json.Unmarshal(body, &tokenInfo); err != nil {
		return nil, fmt.Errorf("failed to parse token info: %w", err)
	}

	// Extract user information from tokeninfo
	user := &entities.User{
		ID:       getStringFromMap(tokenInfo, "user_id"),
		Email:    getStringFromMap(tokenInfo, "email"),
		Verified: getBoolFromMap(tokenInfo, "verified_email"),
	}

	// If we have a user_id but no email, try to get user profile from Google+ API
	if user.ID != "" {
		// Try to get additional user info using the access token
		userInfo, err := getUserInfoFromGoogle(accessToken)
		if err == nil {
			user.Email = userInfo.Email
			user.Name = userInfo.Name
			user.Picture = userInfo.Picture
		}
	}

	// Assign role based on user information
	user.Role = getRoleService().GetUserRole(user)

	return user, nil
}

// getUserInfoFromGoogle fetches additional user information from Google's userinfo endpoint
func getUserInfoFromGoogle(accessToken string) (*entities.User, error) {
	url := "https://www.googleapis.com/oauth2/v2/userinfo"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read user info response: %w", err)
	}

	var userInfo map[string]any
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	return &entities.User{
		ID:       getStringFromMap(userInfo, "id"),
		Email:    getStringFromMap(userInfo, "email"),
		Name:     getStringFromMap(userInfo, "name"),
		Picture:  getStringFromMap(userInfo, "picture"),
		Verified: getBoolFromMap(userInfo, "verified_email"),
		Groups:   []string{}, // TODO: Implement group extraction based on your specific requirements
	}, nil
}

// GetUserFromContext extracts the user from the request context
func GetUserFromContext(ctx context.Context) (*entities.User, bool) {
	user, ok := ctx.Value(UserKey).(*entities.User)
	return user, ok
}

// Helper functions to safely extract values from map[string]interface{}
func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getBoolFromMap(m map[string]interface{}, key string) bool {
	if val, ok := m[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
		if str, ok := val.(string); ok {
			return str == "true"
		}
	}
	return false
}
