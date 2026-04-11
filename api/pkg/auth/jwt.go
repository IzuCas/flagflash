package auth

import (
	"errors"
	"os"
	"time"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

func getSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("auth: JWT_SECRET environment variable must be set")
	}
	return []byte(secret)
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// FlagFlashClaims represents JWT claims for FlagFlash users
type FlagFlashClaims struct {
	UserID   uuid.UUID       `json:"user_id"`
	TenantID uuid.UUID       `json:"tenant_id"`
	Email    string          `json:"email"`
	Name     string          `json:"name"`
	Role     entity.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken creates a signed JWT for the given username (legacy auth)
func GenerateToken(username string) (string, error) {
	expiry := time.Duration(24) * time.Hour

	claims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "flagflash",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getSecret())
}

// GenerateJWT creates a signed JWT for FlagFlash users
func GenerateJWT(userClaims *entity.UserClaims, secret string, expiry time.Duration) (string, error) {
	claims := FlagFlashClaims{
		UserID:   userClaims.UserID,
		TenantID: userClaims.TenantID,
		Email:    userClaims.Email,
		Name:     userClaims.Name,
		Role:     userClaims.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "flagflash",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateJWT parses and validates a FlagFlash JWT string
func ValidateJWT(tokenStr string, secret string) (*FlagFlashClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &FlagFlashClaims{}, func(t *jwt.Token) (interface{}, error) {
		if m, ok := t.Method.(*jwt.SigningMethodHMAC); !ok || m != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*FlagFlashClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

// ValidateToken parses and validates a JWT string, returning the claims
func ValidateToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if m, ok := t.Method.(*jwt.SigningMethodHMAC); !ok || m != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return getSecret(), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
