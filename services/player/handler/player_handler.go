package handler

import (
	"context"

	"github.com/xrekson/game-backend/proto/player"
	"github.com/xrekson/game-backend/services/player/service"
)

type PlayerHandler struct {
	playerpb.UnimplementedPlayerServiceServer
	service service.PlayerService
}

func NewPlayerHandler(service service.PlayerService) *PlayerHandler {
	return &PlayerHandler{service: service}
}

func (h *PlayerHandler) GetProfile(ctx context.Context, req *playerpb.GetProfileRequest) (*playerpb.GetProfileResponse, error) {
	profile, err := h.service.GetProfile(ctx, req.PlayerId)
	if err != nil {
		return nil, err
	}

	return &playerpb.GetProfileResponse{
		Profile: &playerpb.PlayerProfile{
			PlayerId:  profile.PlayerID,
			Username:  profile.Username,
			Level:     profile.Level,
			Xp:        profile.XP,
			AvatarUrl: profile.AvatarURL,
			CreatedAt: profile.CreatedAt.Unix(),
		},
	}, nil
}

func (h *PlayerHandler) UpdateProfile(ctx context.Context, req *playerpb.UpdateProfileRequest) (*playerpb.UpdateProfileResponse, error) {
	profile, err := h.service.UpdateProfile(ctx, req.PlayerId, req.AvatarUrl)
	if err != nil {
		return nil, err
	}

	return &playerpb.UpdateProfileResponse{
		Profile: &playerpb.PlayerProfile{
			PlayerId:  profile.PlayerID,
			Username:  profile.Username,
			Level:     profile.Level,
			Xp:        profile.XP,
			AvatarUrl: profile.AvatarURL,
			CreatedAt: profile.CreatedAt.Unix(),
		},
	}, nil
}
