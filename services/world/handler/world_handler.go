package handler

import (
	"context"

	"github.com/xrekson/game-backend/proto/world"
	"github.com/xrekson/game-backend/services/world/model"
	"github.com/xrekson/game-backend/services/world/service"
)

type WorldHandler struct {
	worldpb.UnimplementedWorldServiceServer
	service service.WorldService
}

func NewWorldHandler(service service.WorldService) *WorldHandler {
	return &WorldHandler{service: service}
}

func (h *WorldHandler) GetMapLayout(ctx context.Context, req *worldpb.GetMapLayoutRequest) (*worldpb.GetMapLayoutResponse, error) {
	layout, err := h.service.GetMapLayout(ctx, req.MapId)
	if err != nil {
		return nil, err
	}

	return &worldpb.GetMapLayoutResponse{
		Map: &worldpb.MapLayout{
			MapId:      layout.MapID,
			Name:       layout.Name,
			Width:      layout.Width,
			Height:     layout.Height,
			IsSafeZone: layout.IsSafeZone,
			SpawnPoint: &worldpb.Position{X: layout.SpawnPoint.X, Y: layout.SpawnPoint.Y},
		},
	}, nil
}

func (h *WorldHandler) ValidateMovement(ctx context.Context, req *worldpb.ValidateMovementRequest) (*worldpb.ValidateMovementResponse, error) {
	pos := model.Position{}
	if req.NewPosition != nil {
		pos.X = req.NewPosition.X
		pos.Y = req.NewPosition.Y
	}

	validated, isSafe, err := h.service.ValidateMovement(ctx, req.PlayerId, req.MapId, pos)
	if err != nil {
		return nil, err
	}

	return &worldpb.ValidateMovementResponse{
		Valid:             true,
		ValidatedPosition: &worldpb.Position{X: validated.X, Y: validated.Y},
		IsSafeZone:        isSafe,
	}, nil
}

func (h *WorldHandler) ResolveCombat(ctx context.Context, req *worldpb.ResolveCombatRequest) (*worldpb.ResolveCombatResponse, error) {
	defeated, dealt, remainingHP, gold, xp, err := h.service.ResolveCombat(ctx, req.PlayerId, req.EnemyId, req.Damage)
	if err != nil {
		return nil, err
	}

	return &worldpb.ResolveCombatResponse{
		EnemyDefeated:    defeated,
		DamageDealt:      dealt,
		RemainingEnemyHp: remainingHP,
		RewardGold:       gold,
		RewardXp:         xp,
	}, nil
}
