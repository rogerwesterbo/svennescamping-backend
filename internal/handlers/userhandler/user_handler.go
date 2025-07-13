package userhandler

import (
	"net/http"

	"github.com/rogerwesterbo/svennescamping-backend/internal/middlewares"
	"github.com/rogerwesterbo/svennescamping-backend/pkg/helpers/httphelpers"
	"go.uber.org/zap"
)

// UserHandler returns the authenticated user's information
func UserHandler(logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract user from context (set by auth middleware)
		user, ok := middlewares.GetUserFromContext(r.Context())
		if !ok {
			logger.Error("User not found in context",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
			)
			httphelpers.RespondWithJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "User information not available",
			})
			return
		}

		logger.Info("User info requested",
			zap.String("userID", user.ID),
			zap.String("email", user.Email),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)

		// Return user information
		response := user

		err := httphelpers.RespondWithJSON(w, http.StatusOK, response)
		if err != nil {
			logger.Error("Failed to send user response",
				zap.Error(err),
				zap.String("userID", user.ID),
			)
			return
		}

		logger.Info("User info response sent successfully",
			zap.String("userID", user.ID),
			zap.String("email", user.Email),
		)
	}
}
