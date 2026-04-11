package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

// ─── WebSocket wire types (internal) ────────────────────────────

type wsMessageType string

const (
	wsMsgFlagUpdate  wsMessageType = "flag_update"
	wsMsgFlagDelete  wsMessageType = "flag_delete"
	wsMsgFlagsSync   wsMessageType = "flags_sync"
	wsMsgSubscribe   wsMessageType = "subscribe"
	wsMsgUnsubscribe wsMessageType = "unsubscribe"
	wsMsgPing        wsMessageType = "ping"
	wsMsgPong        wsMessageType = "pong"
	wsMsgError       wsMessageType = "error"
)

type wsFlag struct {
	ID      string          `json:"id"`
	Key     string          `json:"key"`
	Name    string          `json:"name"`
	Type    string          `json:"type"`
	Enabled bool            `json:"enabled"`
	Value   json.RawMessage `json:"value"`
	Version int             `json:"version"`
}

type wsMessage struct {
	Type  wsMessageType `json:"type"`
	Flag  *wsFlag       `json:"flag,omitempty"`
	Flags []*wsFlag     `json:"flags,omitempty"`
	Key   string        `json:"flag_key,omitempty"`
	Error string        `json:"error,omitempty"`
}

// ─── WebSocket reconnect loop ─────────────────────────────────────

func (c *Client) wsLoop(ctx context.Context) {
	delay := reconnectBaseDelay
	for {
		select {
		case <-c.stopCh:
			return
		case <-ctx.Done():
			return
		default:
		}

		if err := c.wsConnect(ctx); err != nil {
			log.Printf("flagflash: ws disconnected: %v — reconnecting in %s", err, delay)
			select {
			case <-time.After(delay):
			case <-c.stopCh:
				return
			case <-ctx.Done():
				return
			}
			if delay *= reconnectMultiplier; delay > reconnectMaxDelay {
				delay = reconnectMaxDelay
			}
		} else {
			delay = reconnectBaseDelay
		}
	}
}

func (c *Client) wsConnect(ctx context.Context) error {
	hdr := http.Header{"X-API-Key": {c.apiKey}}
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, c.wsURL, hdr)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	// Re-bootstrap on every (re)connect to not miss updates that happened while disconnected
	if err := c.bootstrapCache(ctx); err != nil {
		log.Printf("flagflash: re-bootstrap failed (using stale cache): %v", err)
	}

	for {
		select {
		case <-c.stopCh:
			_ = conn.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return nil
		case <-ctx.Done():
			return nil
		default:
		}

		conn.SetReadDeadline(time.Now().Add(90 * time.Second))
		_, data, err := conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("read: %w", err)
		}

		// Server batches messages with newlines
		for _, chunk := range bytes.Split(data, []byte{'\n'}) {
			if chunk = bytes.TrimSpace(chunk); len(chunk) > 0 {
				c.handleWSMessage(chunk)
			}
		}
	}
}

func (c *Client) handleWSMessage(data []byte) {
	var msg wsMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("[flagflash] WARNING: received malformed WebSocket message (%d bytes): %v", len(data), err)
		return
	}
	switch msg.Type {
	case wsMsgFlagUpdate:
		if msg.Flag != nil {
			c.setInCache(msg.Flag.Key, wsFlagToResult(msg.Flag))
		}
	case wsMsgFlagDelete:
		if msg.Key != "" {
			c.deleteFromCache(msg.Key)
		}
	case wsMsgFlagsSync:
		for _, f := range msg.Flags {
			c.setInCache(f.Key, wsFlagToResult(f))
		}
	}
}

func wsFlagToResult(f *wsFlag) *EvalResult {
	val := f.Value
	if len(val) == 0 {
		if f.Enabled {
			val = json.RawMessage(`true`)
		} else {
			val = json.RawMessage(`false`)
		}
	}
	return &EvalResult{
		FlagKey:   f.Key,
		Value:     val,
		Enabled:   f.Enabled,
		Version:   f.Version,
		FromCache: true,
	}
}
