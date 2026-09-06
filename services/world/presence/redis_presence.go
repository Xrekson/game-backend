package presence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/xrekson/game-backend/pkg/logger"
	"github.com/xrekson/game-backend/pkg/redis"
	"github.com/xrekson/game-backend/services/world/model"
)

type PresenceManager struct {
	redis *redis.Client
}

func NewPresenceManager(r *redis.Client) *PresenceManager {
	return &PresenceManager{redis: r}
}

func (p *PresenceManager) UpdatePresence(ctx context.Context, state *model.PlayerState) error {
	if p.redis == nil {
		return nil
	}

	key := fmt.Sprintf("presence:player:%s", state.PlayerID)
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	// Set Redis location key with TTL
	if err := p.redis.Set(ctx, key, data, 5*time.Minute).Err(); err != nil {
		logger.Error("Failed setting presence in Redis: %v", err)
	}

	// Publish to map channel for real-time presence fanout
	channel := fmt.Sprintf("presence:map:%s", state.MapID)
	return p.redis.Publish(ctx, channel, data).Err()
}

func (p *PresenceManager) GetPresence(ctx context.Context, playerID string) (*model.PlayerState, error) {
	if p.redis == nil {
		return nil, nil
	}

	key := fmt.Sprintf("presence:player:%s", playerID)
	data, err := p.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	state := &model.PlayerState{}
	if err := json.Unmarshal(data, state); err != nil {
		return nil, err
	}
	return state, nil
}
