package sdk

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	defaultTimeout      = 5 * time.Second
	reconnectBaseDelay  = 1 * time.Second
	reconnectMaxDelay   = 30 * time.Second
	reconnectMultiplier = 2
	// maxCacheSize caps the in-memory flag cache to prevent unbounded growth.
	maxCacheSize = 10_000
)

// Client is the FlagFlash SDK client.
type Client struct {
	apiKey  string
	baseURL string
	wsURL   string
	http    *http.Client

	// local flag cache: key → *EvalResult
	cache   map[string]*EvalResult
	cacheMu sync.RWMutex

	// lifecycle
	connected bool
	closeOnce sync.Once
	stopCh    chan struct{}
}

// New creates a new FlagFlash client.
//
// apiKey is the API key generated in the FlagFlash dashboard.
// serverURL is the base URL, e.g. "http://flagflash.example.com".
func New(apiKey, serverURL string, opts ...Option) *Client {
	if apiKey == "" {
		panic("flagflash: apiKey must not be empty")
	}
	serverURL = strings.TrimRight(serverURL, "/")
	if !strings.HasPrefix(serverURL, "https://") {
		log.Printf("[flagflash] WARNING: connecting to non-HTTPS URL %q — credentials and flag data will be transmitted in plain text", serverURL)
	}

	wsScheme := "ws"
	httpScheme := "http"
	host := serverURL
	if strings.HasPrefix(serverURL, "https://") {
		wsScheme = "wss"
		httpScheme = "https"
		host = strings.TrimPrefix(serverURL, "https://")
	} else {
		host = strings.TrimPrefix(serverURL, "http://")
	}

	c := &Client{
		apiKey:  apiKey,
		baseURL: httpScheme + "://" + host,
		wsURL:   wsScheme + "://" + host + "/api/v1/flagflash/sdk/ws",
		http:    &http.Client{},
		cache:   make(map[string]*EvalResult),
		stopCh:  make(chan struct{}),
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Connect bootstraps the local cache with an initial HTTP fetch and then
// opens a persistent WebSocket connection that keeps the cache in sync.
// It blocks until the initial fetch is done; the WS goroutine runs in background.
//
// Call Close() or cancel ctx to stop the background goroutine.
func (c *Client) Connect(ctx context.Context) error {
	if err := c.bootstrapCache(ctx); err != nil {
		return fmt.Errorf("flagflash: bootstrap: %w", err)
	}
	go c.wsLoop(ctx)
	c.connected = true
	return nil
}

// Close terminates the background WebSocket goroutine.
func (c *Client) Close() {
	c.closeOnce.Do(func() { close(c.stopCh) })
}

// Connected reports whether Connect completed successfully.
func (c *Client) Connected() bool { return c.connected }

// ─── cache helpers ────────────────────────────────────────────────

func (c *Client) getFromCache(key string) (*EvalResult, bool) {
	c.cacheMu.RLock()
	defer c.cacheMu.RUnlock()
	r, ok := c.cache[key]
	return r, ok
}

func (c *Client) setInCache(key string, r *EvalResult) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	// Evict an arbitrary entry if the cache is at capacity and this is a new key.
	if _, exists := c.cache[key]; !exists && len(c.cache) >= maxCacheSize {
		for k := range c.cache {
			delete(c.cache, k)
			break
		}
	}
	c.cache[key] = r
}

func (c *Client) deleteFromCache(key string) {
	c.cacheMu.Lock()
	delete(c.cache, key)
	c.cacheMu.Unlock()
}

func (c *Client) allFromCache() AllFlagsResult {
	c.cacheMu.RLock()
	defer c.cacheMu.RUnlock()
	out := make(AllFlagsResult, len(c.cache))
	for k, v := range c.cache {
		cp := *v
		out[k] = &cp
	}
	return out
}

// ─── bootstrap ───────────────────────────────────────────────────

func (c *Client) bootstrapCache(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var resp getFlagsResp
	if err := c.get(ctx, "/api/v1/flagflash/sdk/flags", &resp); err != nil {
		return err
	}
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	for _, f := range resp.Flags {
		c.cache[f.Key] = &EvalResult{
			FlagKey:   f.Key,
			Value:     f.DefaultValue,
			Enabled:   f.Enabled,
			Version:   f.Version,
			FromCache: true,
		}
	}
	return nil
}
