package service

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/xrekson/game-backend/pkg/errors"
	"github.com/xrekson/game-backend/pkg/events"
	"github.com/xrekson/game-backend/pkg/kafka"
	"github.com/xrekson/game-backend/services/world/model"
	"github.com/xrekson/game-backend/services/world/presence"
	"github.com/xrekson/game-backend/services/world/repository"
)

type WorldService interface {
	GetMapLayout(ctx context.Context, mapID string) (*model.MapLayout, error)
	ValidateMovement(ctx context.Context, playerID, mapID string, targetPos model.Position) (*model.Position, bool, error)
	ResolveCombat(ctx context.Context, playerID, enemyID string, damage float64) (bool, float64, float64, float64, int64, error)
}

type worldService struct {
	mapRepo         repository.MapRepository
	presenceManager *presence.PresenceManager
	producer        *kafka.Producer
}

func NewWorldService(
	mapRepo repository.MapRepository,
	presenceMgr *presence.PresenceManager,
	producer *kafka.Producer,
) WorldService {
	return &worldService{
		mapRepo:         mapRepo,
		presenceManager: presenceMgr,
		producer:        producer,
	}
}

func (s *worldService) GetMapLayout(ctx context.Context, mapID string) (*model.MapLayout, error) {
	layout, err := s.mapRepo.GetMapLayout(ctx, mapID)
	if err != nil {
		return nil, errors.NotFound("map not found", err)
	}
	return layout, nil
}

func (s *worldService) ValidateMovement(ctx context.Context, playerID, mapID string, targetPos model.Position) (*model.Position, bool, error) {
	layout, err := s.GetMapLayout(ctx, mapID)
	if err != nil {
		return nil, false, err
	}

	// Clamp coordinates within 2D map boundaries
	validated := targetPos
	if validated.X < 0 {
		validated.X = 0
	} else if validated.X > layout.Width {
		validated.X = layout.Width
	}

	if validated.Y < 0 {
		validated.Y = 0
	} else if validated.Y > layout.Height {
		validated.Y = layout.Height
	}

	// Update presence state in Redis
	state := &model.PlayerState{
		PlayerID:   playerID,
		MapID:      mapID,
		Position:   validated,
		HP:         100.0,
		MaxHP:      100.0,
		IsSafeZone: layout.IsSafeZone,
	}
	if s.presenceManager != nil {
		_ = s.presenceManager.UpdatePresence(ctx, state)
	}

	return &validated, layout.IsSafeZone, nil
}

func (s *worldService) ResolveCombat(ctx context.Context, playerID, enemyID string, damage float64) (bool, float64, float64, float64, int64, error) {
	if damage <= 0 {
		damage = 10.0
	}

	enemyMaxHP := 100.0
	remainingHP := enemyMaxHP - damage
	defeated := remainingHP <= 0

	var rewardGold float64 = 0
	var rewardXP int64 = 0

	if defeated {
		remainingHP = 0
		rewardGold = float64(rand.Intn(50) + 20)
		rewardXP = int64(rand.Intn(100) + 50)
	}

	if s.producer != nil {
		_ = s.producer.Publish(ctx, events.TopicWorldCombat, playerID, events.CombatEvent{
			PlayerID:    playerID,
			DungeonID:   "dungeon_1",
			EnemyID:     enemyID,
			DamageDealt: damage,
			PlayerHP:    100.0,
			EnemyHP:     remainingHP,
			Result:      fmt.Sprintf("%v", defeated),
		})
	}

	return defeated, damage, remainingHP, rewardGold, rewardXP, nil
}
