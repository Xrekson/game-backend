package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/xrekson/game-backend/pkg/postgres"
	"github.com/xrekson/game-backend/services/economy/model"
)

type EconomyRepository interface {
	GetBalance(ctx context.Context, playerID string) (*model.PlayerBalance, error)
	CreateBalance(ctx context.Context, balance *model.PlayerBalance) error
	UpdateBalance(ctx context.Context, balance *model.PlayerBalance) error
	GetUpgrades(ctx context.Context, playerID string) ([]*model.PlayerUpgrade, error)
	GetUpgrade(ctx context.Context, playerID, upgradeID string) (*model.PlayerUpgrade, error)
	SaveUpgrade(ctx context.Context, upgrade *model.PlayerUpgrade) error
	DeductCurrency(ctx context.Context, playerID string, amount float64) (float64, error)
	AddCurrency(ctx context.Context, playerID string, amount float64) (float64, error)
}

type postgresEconomyRepository struct {
	db *postgres.DB
}

func NewEconomyRepository(db *postgres.DB) EconomyRepository {
	return &postgresEconomyRepository{db: db}
}

func (r *postgresEconomyRepository) GetBalance(ctx context.Context, playerID string) (*model.PlayerBalance, error) {
	query := `SELECT player_id, balance, base_income_per_sec, prestige_level, prestige_multiplier, last_claim_time, updated_at 
			  FROM player_balances WHERE player_id = $1`
	b := &model.PlayerBalance{}
	err := r.db.QueryRowContext(ctx, query, playerID).Scan(
		&b.PlayerID, &b.Balance, &b.BaseIncomePerSec, &b.PrestigeLevel, &b.PrestigeMultiplier, &b.LastClaimTime, &b.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return b, nil
}

func (r *postgresEconomyRepository) CreateBalance(ctx context.Context, b *model.PlayerBalance) error {
	query := `INSERT INTO player_balances (player_id, balance, base_income_per_sec, prestige_level, prestige_multiplier, last_claim_time, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query, b.PlayerID, b.Balance, b.BaseIncomePerSec, b.PrestigeLevel, b.PrestigeMultiplier, b.LastClaimTime, time.Now())
	return err
}

func (r *postgresEconomyRepository) UpdateBalance(ctx context.Context, b *model.PlayerBalance) error {
	query := `UPDATE player_balances SET balance = $1, base_income_per_sec = $2, prestige_level = $3, prestige_multiplier = $4, last_claim_time = $5, updated_at = $6 WHERE player_id = $7`
	_, err := r.db.ExecContext(ctx, query, b.Balance, b.BaseIncomePerSec, b.PrestigeLevel, b.PrestigeMultiplier, b.LastClaimTime, time.Now(), b.PlayerID)
	return err
}

func (r *postgresEconomyRepository) GetUpgrades(ctx context.Context, playerID string) ([]*model.PlayerUpgrade, error) {
	query := `SELECT player_id, upgrade_id, level, updated_at FROM player_upgrades WHERE player_id = $1`
	rows, err := r.db.QueryContext(ctx, query, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Err() != nil {
		return nil, rows.Err()
	}
	var upgrades []*model.PlayerUpgrade
	for rows.Next() {
		u := &model.PlayerUpgrade{}
		if err := rows.Scan(&u.PlayerID, &u.UpgradeID, &u.Level, &u.UpdatedAt); err != nil {
			return nil, err
		}
		upgrades = append(upgrades, u)
	}
	return upgrades, nil
}

func (r *postgresEconomyRepository) GetUpgrade(ctx context.Context, playerID, upgradeID string) (*model.PlayerUpgrade, error) {
	query := `SELECT player_id, upgrade_id, level, updated_at FROM player_upgrades WHERE player_id = $1 AND upgrade_id = $2`
	u := &model.PlayerUpgrade{}
	err := r.db.QueryRowContext(ctx, query, playerID, upgradeID).Scan(&u.PlayerID, &u.UpgradeID, &u.Level, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

func (r *postgresEconomyRepository) SaveUpgrade(ctx context.Context, u *model.PlayerUpgrade) error {
	query := `INSERT INTO player_upgrades (player_id, upgrade_id, level, updated_at)
			  VALUES ($1, $2, $3, $4)
			  ON CONFLICT (player_id, upgrade_id) DO UPDATE SET level = EXCLUDED.level, updated_at = EXCLUDED.updated_at`
	_, err := r.db.ExecContext(ctx, query, u.PlayerID, u.UpgradeID, u.Level, time.Now())
	return err
}

func (r *postgresEconomyRepository) DeductCurrency(ctx context.Context, playerID string, amount float64) (float64, error) {
	query := `UPDATE player_balances SET balance = balance - $1, updated_at = $2 WHERE player_id = $3 AND balance >= $1 RETURNING balance`
	var newBalance float64
	err := r.db.QueryRowContext(ctx, query, amount, time.Now(), playerID).Scan(&newBalance)
	if err != nil {
		return 0, err
	}
	return newBalance, nil
}

func (r *postgresEconomyRepository) AddCurrency(ctx context.Context, playerID string, amount float64) (float64, error) {
	query := `UPDATE player_balances SET balance = balance + $1, updated_at = $2 WHERE player_id = $3 RETURNING balance`
	var newBalance float64
	err := r.db.QueryRowContext(ctx, query, amount, time.Now(), playerID).Scan(&newBalance)
	if err != nil {
		return 0, err
	}
	return newBalance, nil
}
