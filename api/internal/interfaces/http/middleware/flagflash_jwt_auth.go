package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/IzuCas/flagflash/internal/application/service"
	"github.com/IzuCas/flagflash/internal/domain/repository"
	"github.com/IzuCas/flagflash/pkg/auth"
	"github.com/google/uuid"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	TenantIDKey contextKey = "tenant_id"
	ClaimsKey   contextKey = "claims"
)

// FlagFlashJWTAuth creates middleware for JWT authentication for FlagFlash dashboard
func FlagFlashJWTAuth(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error": "Authorization header required"}`, http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, `{"error": "Invalid authorization format"}`, http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")

			// Validate token using auth package
			claims, err := auth.ValidateJWT(token, authService.GetJWTSecret())
			if err != nil {
				http.Error(w, `{"error": "Invalid token"}`, http.StatusUnauthorized)
				return
			}

			// Add user info to context
			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDKey, claims.UserID.String())
			ctx = context.WithValue(ctx, TenantIDKey, claims.TenantID.String())
			ctx = context.WithValue(ctx, ClaimsKey, claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalJWTAuth creates middleware that optionally validates JWT if present
func OptionalJWTAuth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				claims, err := auth.ValidateJWT(token, jwtSecret)
				if err == nil {
					ctx := r.Context()
					ctx = context.WithValue(ctx, UserIDKey, claims.UserID.String())
					ctx = context.WithValue(ctx, TenantIDKey, claims.TenantID.String())
					ctx = context.WithValue(ctx, ClaimsKey, claims)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) string {
	if v := ctx.Value(UserIDKey); v != nil {
		return v.(string)
	}
	return ""
}

// GetTenantIDFromContext extracts tenant ID from context
func GetTenantIDFromContext(ctx context.Context) string {
	if v := ctx.Value(TenantIDKey); v != nil {
		return v.(string)
	}
	return ""
}

// GetClaimsFromContext extracts claims from context
func GetClaimsFromContext(ctx context.Context) *auth.FlagFlashClaims {
	if v := ctx.Value(ClaimsKey); v != nil {
		return v.(*auth.FlagFlashClaims)
	}
	return nil
}

// TenantAccessValidator creates middleware that validates user has access to the tenant
// It validates the tenant_id from the JWT claims against the user's memberships
func TenantAccessValidator(membershipRepo repository.UserTenantMembershipRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaimsFromContext(r.Context())
			if claims == nil {
				http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
				return
			}

			// Validate user has access to the tenant in the JWT
			hasAccess, err := membershipRepo.ExistsByUserAndTenant(r.Context(), claims.UserID, claims.TenantID)
			if err != nil || !hasAccess {
				http.Error(w, `{"error": "Access denied to this tenant"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ValidateTenantFromPath creates middleware that validates the tenant_id in the path
// matches the tenant in the JWT claims, ensuring users can only access their own tenant's resources
func ValidateTenantFromPath(paramName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaimsFromContext(r.Context())
			if claims == nil {
				http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
				return
			}

			// Extract tenant_id from path using chi's URL params
			pathTenantID := r.PathValue(paramName)
			if pathTenantID == "" {
				// No tenant in path, continue
				next.ServeHTTP(w, r)
				return
			}

			tenantUUID, err := uuid.Parse(pathTenantID)
			if err != nil {
				http.Error(w, `{"error": "Invalid tenant ID"}`, http.StatusBadRequest)
				return
			}

			// Validate the path tenant matches the JWT tenant
			if tenantUUID != claims.TenantID {
				http.Error(w, `{"error": "Access denied - tenant mismatch"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
