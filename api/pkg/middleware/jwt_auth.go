package middleware

import (
	"net/http"
	"strings"

	"github.com/IzuCas/flagflash/pkg/auth"
)

// JWTAuth validates the Authorization: Bearer <token> header.
// Returns 401 if missing or invalid, 403 if expired.
func JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"title":"Unauthorized","status":401,"detail":"Missing Authorization header"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			http.Error(w, `{"title":"Unauthorized","status":401,"detail":"Invalid Authorization header format"}`, http.StatusUnauthorized)
			return
		}

		claims, err := auth.ValidateToken(parts[1])
		if err != nil {
			if err == auth.ErrExpiredToken {
				http.Error(w, `{"title":"Forbidden","status":403,"detail":"Token has expired"}`, http.StatusForbidden)
				return
			}
			http.Error(w, `{"title":"Unauthorized","status":401,"detail":"Invalid token"}`, http.StatusUnauthorized)
			return
		}

		// Reject tokens issued before the last password change
		creds := auth.LoadCredentials()
		if !creds.TokensValidAfter.IsZero() && claims.IssuedAt != nil {
			if claims.IssuedAt.Time.Before(creds.TokensValidAfter) {
				http.Error(w, `{"title":"Unauthorized","status":401,"detail":"Token has been invalidated"}`, http.StatusUnauthorized)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
