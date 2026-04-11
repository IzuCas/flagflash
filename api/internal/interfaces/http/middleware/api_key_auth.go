package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/IzuCas/flagflash/internal/application/service"
	"github.com/google/uuid"
)

// APIKeyAuth creates middleware for API key authentication
func APIKeyAuth(apiKeyService *service.APIKeyService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get API key from header
			apiKey := r.Header.Get("X-API-Key")
			if apiKey == "" {
				// Try Authorization header with Bearer prefix
				auth := r.Header.Get("Authorization")
				if strings.HasPrefix(auth, "Bearer ") {
					apiKey = strings.TrimPrefix(auth, "Bearer ")
				}
			}

			if apiKey == "" {
				http.Error(w, `{"error": "API key required"}`, http.StatusUnauthorized)
				return
			}

			// Validate API key
			keyDetails, err := apiKeyService.ValidateAPIKey(r.Context(), apiKey)
			if err != nil {
				http.Error(w, `{"error": "Invalid API key"}`, http.StatusUnauthorized)
				return
			}

			// Check if key is active and not expired
			if !keyDetails.Active {
				http.Error(w, `{"error": "API key is inactive"}`, http.StatusUnauthorized)
				return
			}

			// Require an environment to be bound to this key for SDK access
			if keyDetails.EnvironmentID == nil {
				http.Error(w, `{"error": "API key has no environment assigned"}`, http.StatusUnauthorized)
				return
			}

			// Add key details to context (dereference pointer types so handlers can
			// do plain uuid.UUID type assertions without worrying about nil pointers)
			ctx := r.Context()
			ctx = context.WithValue(ctx, "api_key_id", keyDetails.ID)
			ctx = context.WithValue(ctx, "tenant_id", keyDetails.TenantID)
			ctx = context.WithValue(ctx, "environment_id", *keyDetails.EnvironmentID)
			ctx = context.WithValue(ctx, "permissions", keyDetails.Permissions)

			// Update last used timestamp asynchronously
			go func(id uuid.UUID) {
				_ = apiKeyService.UpdateLastUsed(context.Background(), id)
			}(keyDetails.ID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePermission creates middleware that checks for required permission
func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			permissions, ok := r.Context().Value("permissions").([]string)
			if !ok {
				http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
				return
			}

			// Check if admin or has required permission
			for _, p := range permissions {
				if p == "admin" || p == permission {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, `{"error": "Insufficient permissions"}`, http.StatusForbidden)
		})
	}
}
