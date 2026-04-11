package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ─── HTTP helpers ─────────────────────────────────────────────────

func (c *Client) post(ctx context.Context, path string, body, dest any) error {
	buf, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("flagflash: marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(buf))
	if err != nil {
		return fmt.Errorf("flagflash: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)
	return c.do(req, dest)
}

func (c *Client) get(ctx context.Context, path string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("flagflash: build request: %w", err)
	}
	req.Header.Set("X-API-Key", c.apiKey)
	return c.do(req, dest)
}

func (c *Client) do(req *http.Request, dest any) error {
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("flagflash: http: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("flagflash: read response: %w", err)
	}
	if resp.StatusCode >= 400 {
		var apiErr struct {
			Title  string `json:"title"`
			Detail string `json:"detail"`
		}
		_ = json.Unmarshal(raw, &apiErr)
		msg := apiErr.Detail
		if msg == "" {
			msg = apiErr.Title
		}
		if msg == "" {
			msg = string(raw)
		}
		return fmt.Errorf("flagflash: server error %d: %s", resp.StatusCode, msg)
	}
	if dest != nil {
		if err := json.Unmarshal(raw, dest); err != nil {
			return fmt.Errorf("flagflash: decode response: %w", err)
		}
	}
	return nil
}
