package service

import (
	"context"
	"time"

	appErrors "github.com/xrekson/game-backend/pkg/errors"
	"github.com/xrekson/game-backend/services/player/model"
	"github.com/xrekson/game-backend/services/player/repository"
)

type PlayerService interface {
	GetProfile(ctx context.Context, playerID string) (*model.PlayerProfile, error)
	UpdateProfile(ctx context.Context, playerID, avatarURL string) (*model.PlayerProfile, error)
}

type playerService struct {
	repo repository.PlayerRepository
}

func NewPlayerService(repo repository.PlayerRepository) PlayerService {
	return &playerService{repo: repo}
}

func (s *playerService) GetProfile(ctx context.Context, playerID string) (*model.PlayerProfile, error) {
	profile, err := s.repo.GetProfile(ctx, playerID)
	if err != nil {
		return nil, appErrors.Internal("failed to fetch player profile", err)
	}
	if profile == nil {
		// Initialize profile if not found
		profile = &model.PlayerProfile{
			PlayerID:  playerID,
			Username:  "Player_" + playerID[:8],
			Level:     1,
			XP:        0,
			AvatarURL: "",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := s.repo.CreateProfile(ctx, profile); err != nil {
			return nil, appErrors.Internal("failed to initialize player profile", err)
		}
	}
	return profile, nil
}

func (s *playerService) UpdateProfile(ctx context.Context, playerID, avatarURL string) (*model.PlayerProfile, error) {
	profile, err := s.GetProfile(ctx, playerID)
	if err != nil {
		return nil, err
	}

	profile.AvatarURL = avatarURL
	profile.UpdatedAt = time.Now()

	if err := s.repo.UpdateProfile(ctx, profile); err != nil {
		return nil, appErrors.Internal("failed to update profile", err)
	}

	return profile, nil
}
