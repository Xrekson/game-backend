package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/xrekson/game-backend/pkg/postgres"
	"github.com/xrekson/game-backend/services/player/model"
)

type PlayerRepository interface {
	GetProfile(ctx context.Context, playerID string) (*model.PlayerProfile, error)
	CreateProfile(ctx context.Context, profile *model.PlayerProfile) error
	UpdateProfile(ctx context.Context, profile *model.PlayerProfile) error
}

type postgresPlayerRepository struct {
	db *postgres.DB
}

func NewPlayerRepository(db *postgres.DB) PlayerRepository {
	return &postgresPlayerRepository{db: db}
}

func (r *postgresPlayerRepository) GetProfile(ctx context.Context, playerID string) (*model.PlayerProfile, error) {
	query := `SELECT player_id, username, level, xp, avatar_url, created_at, updated_at FROM player_profiles WHERE player_id = $1`
	p := &model.PlayerProfile{}
	err := r.db.QueryRowContext(ctx, query, playerID).Scan(
		&p.PlayerID, &p.Username, &p.Level, &p.XP, &p.AvatarURL, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return p, nil
}

func (r *postgresPlayerRepository) CreateProfile(ctx context.Context, profile *model.PlayerProfile) error {
	query := `INSERT INTO player_profiles (player_id, username, level, xp, avatar_url, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query, profile.PlayerID, profile.Username, profile.Level, profile.XP, profile.AvatarURL, time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to create player profile: %w", err)
	}
	return nil
}

func (r *postgresPlayerRepository) UpdateProfile(ctx context.Context, profile *model.PlayerProfile) error {
	query := `UPDATE player_profiles SET level = $1, xp = $2, avatar_url = $3, updated_at = $4 WHERE player_id = $5`
	_, err := r.db.ExecContext(ctx, query, profile.Level, profile.XP, profile.AvatarURL, time.Now(), profile.PlayerID)
	return err
}
