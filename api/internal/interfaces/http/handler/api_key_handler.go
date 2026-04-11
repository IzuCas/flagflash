package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/IzuCas/flagflash/internal/application/service"
	"github.com/IzuCas/flagflash/internal/interfaces/http/dto"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

// APIKeyHandler handles API key related HTTP requests
type APIKeyHandler struct {
	service *service.APIKeyService
}

// NewAPIKeyHandler creates a new API key handler
func NewAPIKeyHandler(service *service.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{service: service}
}

// RegisterRoutes registers API key routes
func (h *APIKeyHandler) RegisterRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "createAPIKey",
		Method:      http.MethodPost,
		Path:        "/tenants/{tenant_id}/api-keys",
		Summary:     "Create a new API key",
		Tags:        []string{"API Keys"},
	}, h.CreateAPIKey)

	huma.Register(api, huma.Operation{
		OperationID: "getAPIKey",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}/api-keys/{key_id}",
		Summary:     "Get API key by ID",
		Tags:        []string{"API Keys"},
	}, h.GetAPIKey)

	huma.Register(api, huma.Operation{
		OperationID: "listAPIKeys",
		Method:      http.MethodGet,
		Path:        "/tenants/{tenant_id}/api-keys",
		Summary:     "List API keys for a tenant",
		Tags:        []string{"API Keys"},
	}, h.ListAPIKeys)

	huma.Register(api, huma.Operation{
		OperationID: "revokeAPIKey",
		Method:      http.MethodPost,
		Path:        "/tenants/{tenant_id}/api-keys/{key_id}/revoke",
		Summary:     "Revoke API key",
		Tags:        []string{"API Keys"},
	}, h.RevokeAPIKey)

	huma.Register(api, huma.Operation{
		OperationID: "deleteAPIKey",
		Method:      http.MethodDelete,
		Path:        "/tenants/{tenant_id}/api-keys/{key_id}",
		Summary:     "Delete API key",
		Tags:        []string{"API Keys"},
	}, h.DeleteAPIKey)
}

// CreateAPIKey creates a new API key
func (h *APIKeyHandler) CreateAPIKey(ctx context.Context, req *dto.CreateAPIKeyRequest) (*dto.APIKeyCreatedResponse, error) {
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid tenant ID", err)
	}

	var envID *uuid.UUID
	if req.Body.EnvironmentID != uuid.Nil {
		envID = &req.Body.EnvironmentID
	}
	key, rawKey, err := h.service.CreateAPIKey(ctx, tenantID, envID, req.Body.Name, req.Body.Permissions, req.Body.ExpiresAt)
	if err != nil {
		return nil, huma.Error400BadRequest("Failed to create API key", err)
	}

	resp := &dto.APIKeyCreatedResponse{}
	resp.Body.ID = key.ID
	resp.Body.TenantID = key.TenantID
	resp.Body.EnvironmentID = key.EnvironmentID
	resp.Body.Name = key.Name
	resp.Body.KeyPrefix = key.KeyPrefix
	resp.Body.Permissions = key.Permissions
	resp.Body.Active = key.Active
	resp.Body.ExpiresAt = key.ExpiresAt
	resp.Body.CreatedAt = key.CreatedAt
	resp.Body.Key = rawKey // Only returned on creation

	return resp, nil
}

// GetAPIKeyRequest represents request for getting an API key
type GetAPIKeyRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	KeyID    string `path:"key_id" format:"uuid"`
}

// GetAPIKey retrieves an API key by ID
func (h *APIKeyHandler) GetAPIKey(ctx context.Context, req *GetAPIKeyRequest) (*dto.APIKeyResponse, error) {
	keyID, err := uuid.Parse(req.KeyID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid key ID", err)
	}

	key, err := h.service.GetAPIKey(ctx, keyID)
	if err != nil {
		return nil, huma.Error404NotFound("API key not found", err)
	}

	return h.buildKeyResponse(key), nil
}

// ListAPIKeysRequest represents request for listing API keys
type ListAPIKeysRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	Page     int    `query:"page" default:"1" minimum:"1"`
	Limit    int    `query:"limit" default:"20" minimum:"1" maximum:"100"`
}

// ListAPIKeys lists API keys for a tenant
func (h *APIKeyHandler) ListAPIKeys(ctx context.Context, req *ListAPIKeysRequest) (*dto.APIKeysListResponse, error) {
	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid tenant ID", err)
	}

	keys, err := h.service.ListAPIKeysByTenant(ctx, tenantID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to list API keys", err)
	}

	resp := &dto.APIKeysListResponse{}
	resp.Body.Pagination.Page = req.Page
	resp.Body.Pagination.Limit = req.Limit
	resp.Body.Pagination.Total = int64(len(keys))
	resp.Body.Pagination.TotalPages = (len(keys) + req.Limit - 1) / req.Limit

	// Apply pagination
	start := (req.Page - 1) * req.Limit
	if start >= len(keys) {
		resp.Body.Keys = []dto.APIKeyDTO{}
		return resp, nil
	}
	end := start + req.Limit
	if end > len(keys) {
		end = len(keys)
	}

	resp.Body.Keys = make([]dto.APIKeyDTO, 0, end-start)
	for _, key := range keys[start:end] {
		resp.Body.Keys = append(resp.Body.Keys, dto.APIKeyDTO{
			ID:            key.ID,
			TenantID:      key.TenantID,
			EnvironmentID: key.EnvironmentID,
			Name:          key.Name,
			KeyPrefix:     key.KeyPrefix,
			Permissions:   key.Permissions,
			Active:        key.Active,
			LastUsedAt:    key.LastUsedAt,
			ExpiresAt:     key.ExpiresAt,
			CreatedAt:     key.CreatedAt,
		})
	}

	return resp, nil
}

// RevokeAPIKeyRequest represents request for revoking an API key
type RevokeAPIKeyRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	KeyID    string `path:"key_id" format:"uuid"`
}

// RevokeAPIKey revokes an API key
func (h *APIKeyHandler) RevokeAPIKey(ctx context.Context, req *RevokeAPIKeyRequest) (*dto.APIKeyResponse, error) {
	keyID, err := uuid.Parse(req.KeyID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid key ID", err)
	}

	if err := h.service.RevokeAPIKey(ctx, keyID); err != nil {
		return nil, huma.Error400BadRequest("Failed to revoke API key", err)
	}

	key, err := h.service.GetAPIKey(ctx, keyID)
	if err != nil {
		return nil, huma.Error500InternalServerError("Failed to get updated key", err)
	}

	return h.buildKeyResponse(key), nil
}

// DeleteAPIKeyRequest represents request for deleting an API key
type DeleteAPIKeyRequest struct {
	TenantID string `path:"tenant_id" format:"uuid"`
	KeyID    string `path:"key_id" format:"uuid"`
}

// DeleteAPIKey deletes an API key
func (h *APIKeyHandler) DeleteAPIKey(ctx context.Context, req *DeleteAPIKeyRequest) (*struct{}, error) {
	keyID, err := uuid.Parse(req.KeyID)
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid key ID", err)
	}

	if err := h.service.DeleteAPIKey(ctx, keyID); err != nil {
		return nil, huma.Error400BadRequest("Failed to delete API key", err)
	}

	return &struct{}{}, nil
}

func (h *APIKeyHandler) buildKeyResponse(key *service.APIKeyDetails) *dto.APIKeyResponse {
	resp := &dto.APIKeyResponse{}
	resp.Body.ID = key.ID
	resp.Body.TenantID = key.TenantID
	resp.Body.EnvironmentID = key.EnvironmentID
	resp.Body.Name = key.Name
	resp.Body.KeyPrefix = key.KeyPrefix
	resp.Body.Permissions = key.Permissions
	resp.Body.Active = key.Active
	resp.Body.LastUsedAt = key.LastUsedAt
	resp.Body.ExpiresAt = key.ExpiresAt
	resp.Body.CreatedAt = key.CreatedAt
	return resp
}

// APIKeyDetails is a type alias for service.APIKeyDetails
type APIKeyDetails struct {
	ID            uuid.UUID  `json:"id"`
	TenantID      uuid.UUID  `json:"tenant_id"`
	EnvironmentID *uuid.UUID `json:"environment_id,omitempty"`
	Name          string     `json:"name"`
	KeyPrefix     string     `json:"key_prefix"`
	Permissions   []string   `json:"permissions"`
	Active        bool       `json:"active"`
	LastUsedAt    *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}
