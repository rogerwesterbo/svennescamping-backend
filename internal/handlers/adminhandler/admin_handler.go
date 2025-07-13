package adminhandler

import (
	"encoding/json"
	"net/http"

	"github.com/rogerwesterbo/svennescamping-backend/internal/clients"
	"github.com/rogerwesterbo/svennescamping-backend/internal/middlewares"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/entities"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/helpers/httphelpers"
	"go.uber.org/zap"
)

// RoleAssignmentRequest represents a request to assign a role to a user
type RoleAssignmentRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// ListUsersHandler returns all users and their roles (admin only)
func ListUsersHandler(logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := middlewares.GetUserFromContext(r.Context())
		if !ok {
			logger.Error("User not found in context")
			httphelpers.RespondWithJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Authentication required",
			})
			return
		}

		logger.Info("Admin requested user list",
			zap.String("admin_id", user.ID),
			zap.String("admin_email", user.Email),
		)

		// This is a placeholder - in a real implementation, you'd fetch from a database
		response := map[string]interface{}{
			"message":    "This endpoint would return all users and their roles",
			"admin_user": user,
		}

		err := httphelpers.RespondWithJSON(w, http.StatusOK, response)
		if err != nil {
			logger.Error("Failed to send admin response", zap.Error(err))
		}
	}
}

// AssignRoleHandler allows admins to assign roles to users
func AssignRoleHandler(logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := middlewares.GetUserFromContext(r.Context())
		if !ok {
			logger.Error("User not found in context")
			httphelpers.RespondWithJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Authentication required",
			})
			return
		}

		var req RoleAssignmentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Warn("Invalid role assignment request", zap.Error(err))
			httphelpers.RespondWithJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid request body",
			})
			return
		}

		// Validate the role
		role := entities.Role(req.Role)
		if !role.IsValid() {
			logger.Warn("Invalid role provided",
				zap.String("role", req.Role),
				zap.String("admin_email", user.Email),
			)
			httphelpers.RespondWithJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid role. Valid roles are: admin, user, no_access",
			})
			return
		}

		logger.Info("Admin assigned role",
			zap.String("admin_id", user.ID),
			zap.String("admin_email", user.Email),
			zap.String("target_email", req.Email),
			zap.String("assigned_role", req.Role),
		)

		// In a real implementation, you'd update the database here
		response := map[string]interface{}{
			"message":       "Role assignment would be saved to database",
			"target_email":  req.Email,
			"assigned_role": req.Role,
			"assigned_by":   user.Email,
		}

		err := httphelpers.RespondWithJSON(w, http.StatusOK, response)
		if err != nil {
			logger.Error("Failed to send role assignment response", zap.Error(err))
		}
	}
}

// BackgroundFetcherStatusHandler returns the status of the background transaction fetcher
func BackgroundFetcherStatusHandler(logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := middlewares.GetUserFromContext(r.Context())
		if !ok {
			logger.Error("User not found in context")
			httphelpers.RespondWithJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Authentication required",
			})
			return
		}

		logger.Info("Admin requested background fetcher status",
			zap.String("admin_id", user.ID),
			zap.String("admin_email", user.Email),
		)

		isRunning := clients.IsBackgroundFetchingRunning()

		// Get some cache statistics
		cachedTransactions := clients.Cache.GetTransactions("")

		response := map[string]interface{}{
			"background_fetcher": map[string]interface{}{
				"running":           isRunning,
				"fetch_interval":    "5 minutes",
				"providers_enabled": []string{},
			},
			"cache_stats": map[string]interface{}{
				"total_transactions": len(cachedTransactions),
			},
		}

		// Check which providers are enabled
		var enabledProviders []string
		if clients.StripeClient != nil {
			enabledProviders = append(enabledProviders, "stripe")
		}
		if clients.VippsClient != nil {
			enabledProviders = append(enabledProviders, "vipps")
		}
		if clients.ZettleClient != nil {
			enabledProviders = append(enabledProviders, "zettle")
		}

		response["background_fetcher"].(map[string]interface{})["providers_enabled"] = enabledProviders

		err := httphelpers.RespondWithJSON(w, http.StatusOK, response)
		if err != nil {
			logger.Error("Failed to send background fetcher status response", zap.Error(err))
		}
	}
}
