package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IzuCas/flagflash/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	flagUpdateChannel = "flagflash:flags:update"
	flagDeleteChannel = "flagflash:flags:delete"
)

// FlagMessage represents a flag update/delete message
type FlagMessage struct {
	Type          string              `json:"type"`
	EnvironmentID uuid.UUID           `json:"environment_id"`
	Flag          *entity.FeatureFlag `json:"flag,omitempty"`
	FlagKey       string              `json:"flag_key,omitempty"`
}

// PubSub handles Redis pub/sub for real-time flag updates
type PubSub struct {
	client *Client
}

// NewPubSub creates a new PubSub instance
func NewPubSub(client *Client) *PubSub {
	return &PubSub{client: client}
}

// NewFlagPublisher creates a new FlagPublisher (returns PubSub which implements the interface)
func NewFlagPublisher(client *Client) *PubSub {
	return NewPubSub(client)
}

// PublishFlagUpdate publishes a flag update event
func (p *PubSub) PublishFlagUpdate(ctx context.Context, environmentID uuid.UUID, flag *entity.FeatureFlag) error {
	msg := FlagMessage{
		Type:          "update",
		EnvironmentID: environmentID,
		Flag:          flag,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	return p.client.GetClient().Publish(ctx, flagUpdateChannel, data).Err()
}

// PublishFlagDelete publishes a flag delete event
func (p *PubSub) PublishFlagDelete(ctx context.Context, environmentID uuid.UUID, flagKey string) error {
	msg := FlagMessage{
		Type:          "delete",
		EnvironmentID: environmentID,
		FlagKey:       flagKey,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	return p.client.GetClient().Publish(ctx, flagDeleteChannel, data).Err()
}

// SubscribeAll subscribes to all flag channels
func (p *PubSub) SubscribeAll(ctx context.Context) (<-chan *FlagMessage, func(), error) {
	pubsub := p.client.GetClient().Subscribe(ctx, flagUpdateChannel, flagDeleteChannel)

	// Wait for subscription confirmation
	if _, err := pubsub.Receive(ctx); err != nil {
		pubsub.Close()
		return nil, nil, fmt.Errorf("failed to subscribe: %w", err)
	}

	msgChan := make(chan *FlagMessage, 100)
	cleanup := func() {
		pubsub.Close()
		close(msgChan)
	}

	go func() {
		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				var flagMsg FlagMessage
				if err := json.Unmarshal([]byte(msg.Payload), &flagMsg); err != nil {
					continue
				}
				select {
				case msgChan <- &flagMsg:
				default:
					// Channel full, skip message
				}
			}
		}
	}()

	return msgChan, cleanup, nil
}

// Subscribe subscribes to a specific environment's flag updates
func (p *PubSub) Subscribe(ctx context.Context, environmentID uuid.UUID) (*redis.PubSub, error) {
	channel := fmt.Sprintf("flagflash:env:%s:flags", environmentID.String())
	return p.client.GetClient().Subscribe(ctx, channel), nil
}
