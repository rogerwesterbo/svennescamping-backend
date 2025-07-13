package middlewares

import (
	"net/http"

	"github.com/rogerwesterbo/svennescamping-backend/pkg/entities"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/helpers/httphelpers"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/logger"
	"go.uber.org/zap"
)

// RequireRole creates a middleware that requires a specific role
func RequireRole(requiredRole entities.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := GetUserFromContext(r.Context())
			if !ok {
				logger.Error("User not found in context for role check",
					zap.String("path", r.URL.Path),
					zap.String("required_role", string(requiredRole)),
				)
				httphelpers.RespondWithJSON(w, http.StatusInternalServerError, map[string]string{
					"error": "Authentication required",
				})
				return
			}

			if user.Role != requiredRole {
				logger.Warn("Insufficient permissions",
					zap.String("user_id", user.ID),
					zap.String("user_email", user.Email),
					zap.String("user_role", string(user.Role)),
					zap.String("required_role", string(requiredRole)),
					zap.String("path", r.URL.Path),
				)
				httphelpers.RespondWithJSON(w, http.StatusForbidden, map[string]string{
					"error": "Insufficient permissions",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermission creates a middleware that requires a specific permission
func RequirePermission(requiredPermission entities.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := GetUserFromContext(r.Context())
			if !ok {
				logger.Error("User not found in context for permission check",
					zap.String("path", r.URL.Path),
					zap.String("required_permission", string(requiredPermission)),
				)
				httphelpers.RespondWithJSON(w, http.StatusInternalServerError, map[string]string{
					"error": "Authentication required",
				})
				return
			}

			if !user.HasPermission(requiredPermission) {
				logger.Warn("Insufficient permissions",
					zap.String("user_id", user.ID),
					zap.String("user_email", user.Email),
					zap.String("user_role", string(user.Role)),
					zap.String("required_permission", string(requiredPermission)),
					zap.String("path", r.URL.Path),
				)
				httphelpers.RespondWithJSON(w, http.StatusForbidden, map[string]string{
					"error": "Insufficient permissions",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireMinimumRole creates a middleware that requires at least a minimum role level
func RequireMinimumRole(minimumRole entities.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := GetUserFromContext(r.Context())
			if !ok {
				logger.Error("User not found in context for minimum role check",
					zap.String("path", r.URL.Path),
					zap.String("minimum_role", string(minimumRole)),
				)
				httphelpers.RespondWithJSON(w, http.StatusInternalServerError, map[string]string{
					"error": "Authentication required",
				})
				return
			}

			// Role hierarchy: admin > user > no_access
			userLevel := getRoleLevel(user.Role)
			requiredLevel := getRoleLevel(minimumRole)

			if userLevel < requiredLevel {
				logger.Warn("Insufficient role level",
					zap.String("user_id", user.ID),
					zap.String("user_email", user.Email),
					zap.String("user_role", string(user.Role)),
					zap.String("minimum_role", string(minimumRole)),
					zap.String("path", r.URL.Path),
				)
				httphelpers.RespondWithJSON(w, http.StatusForbidden, map[string]string{
					"error": "Insufficient permissions",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAccess ensures the user has any access (not no_access role)
func RequireAccess() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := GetUserFromContext(r.Context())
			if !ok {
				logger.Error("User not found in context for access check",
					zap.String("path", r.URL.Path),
				)
				httphelpers.RespondWithJSON(w, http.StatusInternalServerError, map[string]string{
					"error": "Authentication required",
				})
				return
			}

			if !user.HasAccess() {
				logger.Warn("User has no access",
					zap.String("user_id", user.ID),
					zap.String("user_email", user.Email),
					zap.String("user_role", string(user.Role)),
					zap.String("path", r.URL.Path),
				)
				httphelpers.RespondWithJSON(w, http.StatusForbidden, map[string]string{
					"error": "Access denied",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getRoleLevel returns the numeric level of a role for hierarchy comparison
func getRoleLevel(role entities.Role) int {
	switch role {
	case entities.RoleAdmin:
		return 3
	case entities.RoleUser:
		return 2
	case entities.RoleNoAccess:
		return 1
	default:
		return 0
	}
}
