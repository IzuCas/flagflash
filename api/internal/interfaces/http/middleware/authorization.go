package middleware

import (
	"context"
	"fmt"

	"github.com/IzuCas/flagflash/pkg/auth"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

// AuthorizationError represents an authorization failure
type AuthorizationError struct {
	message string
}

func (e *AuthorizationError) Error() string {
	return e.message
}

// RequireTenantAccess validates that the authenticated user has access to the specified tenant.
// Returns an error if:
// - User is not authenticated
// - The requested tenant ID doesn't match the user's current tenant from JWT
//
// This MUST be called at the beginning of any handler that accesses tenant-specific resources.
func RequireTenantAccess(ctx context.Context, requestedTenantID string) error {
	claims := GetClaimsFromContext(ctx)
	if claims == nil {
		return huma.Error401Unauthorized("Authentication required")
	}

	// Parse and validate the requested tenant ID
	requestedID, err := uuid.Parse(requestedTenantID)
	if err != nil {
		return huma.Error400BadRequest("Invalid tenant ID format")
	}

	// Verify the user has access to this tenant
	if claims.TenantID != requestedID {
		return huma.Error403Forbidden("Access denied: no permission for this tenant")
	}

	return nil
}

// RequireTenantAccessUUID is like RequireTenantAccess but accepts a uuid.UUID directly
func RequireTenantAccessUUID(ctx context.Context, requestedTenantID uuid.UUID) error {
	claims := GetClaimsFromContext(ctx)
	if claims == nil {
		return huma.Error401Unauthorized("Authentication required")
	}

	if claims.TenantID != requestedTenantID {
		return huma.Error403Forbidden("Access denied: no permission for this tenant")
	}

	return nil
}

// RequireRole validates that the user has the required role level or higher.
// Role hierarchy: owner > admin > member > viewer
func RequireRole(ctx context.Context, requiredRole string) error {
	claims := GetClaimsFromContext(ctx)
	if claims == nil {
		return huma.Error401Unauthorized("Authentication required")
	}

	roleLevel := getRoleLevel(string(claims.Role))
	requiredLevel := getRoleLevel(requiredRole)

	if roleLevel < requiredLevel {
		return huma.Error403Forbidden(fmt.Sprintf("Access denied: requires %s role or higher", requiredRole))
	}

	return nil
}

// RequireAdminOrOwner is a convenience function that requires admin or owner role
func RequireAdminOrOwner(ctx context.Context) error {
	return RequireRole(ctx, "admin")
}

// RequireOwner is a convenience function that requires owner role
func RequireOwner(ctx context.Context) error {
	return RequireRole(ctx, "owner")
}

// ValidateResourceTenant checks if a resource belongs to the user's tenant.
// This should be called after fetching a resource to verify ownership.
func ValidateResourceTenant(ctx context.Context, resourceTenantID uuid.UUID) error {
	claims := GetClaimsFromContext(ctx)
	if claims == nil {
		return huma.Error401Unauthorized("Authentication required")
	}

	if claims.TenantID != resourceTenantID {
		return huma.Error403Forbidden("Access denied: resource belongs to another tenant")
	}

	return nil
}

// GetAuthenticatedUserID returns the authenticated user's ID
func GetAuthenticatedUserID(ctx context.Context) (uuid.UUID, error) {
	claims := GetClaimsFromContext(ctx)
	if claims == nil {
		return uuid.Nil, huma.Error401Unauthorized("Authentication required")
	}
	return claims.UserID, nil
}

// GetAuthenticatedTenantID returns the authenticated user's current tenant ID
func GetAuthenticatedTenantID(ctx context.Context) (uuid.UUID, error) {
	claims := GetClaimsFromContext(ctx)
	if claims == nil {
		return uuid.Nil, huma.Error401Unauthorized("Authentication required")
	}
	return claims.TenantID, nil
}

// GetAuthenticatedClaims returns the full claims from context
func GetAuthenticatedClaims(ctx context.Context) (*auth.FlagFlashClaims, error) {
	claims := GetClaimsFromContext(ctx)
	if claims == nil {
		return nil, huma.Error401Unauthorized("Authentication required")
	}
	return claims, nil
}

// getRoleLevel returns a numeric level for role comparison
func getRoleLevel(role string) int {
	switch role {
	case "owner":
		return 100
	case "admin":
		return 75
	case "member":
		return 50
	case "viewer":
		return 25
	default:
		return 0
	}
}
