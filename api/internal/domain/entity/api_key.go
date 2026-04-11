package entity

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Permission defines API key permissions
type Permission string

const (
	PermissionRead  Permission = "read"
	PermissionWrite Permission = "write"
	PermissionAdmin Permission = "admin"
)

// APIKey represents an API key for SDK authentication
type APIKey struct {
	ID            uuid.UUID  `json:"id"`
	TenantID      uuid.UUID  `json:"tenant_id"`
	EnvironmentID *uuid.UUID `json:"environment_id,omitempty"`
	Name          string     `json:"name"`
	KeyHash       string     `json:"-"` // Never expose the hash
	KeyPrefix     string     `json:"key_prefix"`
	Permissions   []string   `json:"permissions"`
	Active        bool       `json:"active"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	LastUsedAt    *time.Time `json:"last_used_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	RevokedAt     *time.Time `json:"revoked_at,omitempty"`
}

// APIKeyWithPlainKey includes the plain key (only returned on creation)
type APIKeyWithPlainKey struct {
	APIKey
	PlainKey string `json:"key"`
}

// NewAPIKey creates a new API key entity
func NewAPIKey(tenantID uuid.UUID, environmentID *uuid.UUID, name, keyHash, keyPrefix string, permissions []string, expiresAt *time.Time) *APIKey {
	now := time.Now()
	return &APIKey{
		ID:            uuid.New(),
		TenantID:      tenantID,
		EnvironmentID: environmentID,
		Name:          name,
		KeyHash:       keyHash,
		KeyPrefix:     keyPrefix,
		Permissions:   permissions,
		Active:        true,
		ExpiresAt:     expiresAt,
		CreatedAt:     now,
	}
}

// RecordUsage updates the last used timestamp
func (k *APIKey) RecordUsage() {
	now := time.Now()
	k.LastUsedAt = &now
}

// GenerateAPIKey generates a new API key with the format: ff_env_<random>
func GenerateAPIKey(tenantID uuid.UUID, environmentID *uuid.UUID, name string, permissions []string, expiresAt *time.Time) (*APIKeyWithPlainKey, error) {
	// Generate random bytes for the key
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Create the plain key with prefix
	envStr := "all"
	if environmentID != nil {
		envStr = environmentID.String()[:8]
	}
	plainKey := fmt.Sprintf("ff_%s_%s", envStr, hex.EncodeToString(randomBytes))
	keyPrefix := plainKey[:12]

	// Hash the key for storage
	hash := sha256.Sum256([]byte(plainKey))
	keyHash := hex.EncodeToString(hash[:])

	now := time.Now()
	apiKey := &APIKeyWithPlainKey{
		APIKey: APIKey{
			ID:            uuid.New(),
			TenantID:      tenantID,
			EnvironmentID: environmentID,
			Name:          name,
			KeyHash:       keyHash,
			KeyPrefix:     keyPrefix,
			Permissions:   permissions,
			Active:        true,
			ExpiresAt:     expiresAt,
			CreatedAt:     now,
		},
		PlainKey: plainKey,
	}

	return apiKey, nil
}

// HashAPIKey hashes a plain API key for lookup
func HashAPIKey(plainKey string) string {
	hash := sha256.Sum256([]byte(plainKey))
	return hex.EncodeToString(hash[:])
}

// Validate checks if the API key is valid
func (k *APIKey) Validate() error {
	if k.RevokedAt != nil {
		return fmt.Errorf("api key has been revoked")
	}
	if k.ExpiresAt != nil && time.Now().After(*k.ExpiresAt) {
		return fmt.Errorf("api key has expired")
	}
	return nil
}

// Revoke revokes the API key
func (k *APIKey) Revoke() {
	now := time.Now()
	k.RevokedAt = &now
	k.Active = false
}

// IsRevoked checks if the API key is revoked
func (k *APIKey) IsRevoked() bool {
	return k.RevokedAt != nil
}

// IsExpired checks if the API key is expired
func (k *APIKey) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*k.ExpiresAt)
}

// HasPermission checks if the API key has a specific permission
func (k *APIKey) HasPermission(perm string) bool {
	for _, p := range k.Permissions {
		if p == perm || p == string(PermissionAdmin) {
			return true
		}
	}
	return false
}

// UpdateLastUsed updates the last used timestamp
func (k *APIKey) UpdateLastUsed() {
	now := time.Now()
	k.LastUsedAt = &now
}

// Rotate generates a new key while keeping the same metadata
func (k *APIKey) Rotate() (*APIKeyWithPlainKey, error) {
	return GenerateAPIKey(k.TenantID, k.EnvironmentID, k.Name, k.Permissions, k.ExpiresAt)
}

// APIKeyInfo represents minimal info about an API key (for listing)
type APIKeyInfo struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	KeyPrefix   string     `json:"key_prefix"`
	Permissions []string   `json:"permissions"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	IsRevoked   bool       `json:"is_revoked"`
}

// ToInfo converts APIKey to APIKeyInfo
func (k *APIKey) ToInfo() *APIKeyInfo {
	return &APIKeyInfo{
		ID:          k.ID,
		Name:        k.Name,
		KeyPrefix:   k.KeyPrefix,
		Permissions: k.Permissions,
		ExpiresAt:   k.ExpiresAt,
		LastUsedAt:  k.LastUsedAt,
		CreatedAt:   k.CreatedAt,
		IsRevoked:   k.IsRevoked(),
	}
}
