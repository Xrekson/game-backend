package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/xrekson/game-backend/pkg/errors"
	"github.com/xrekson/game-backend/pkg/events"
	"github.com/xrekson/game-backend/pkg/kafka"
	"github.com/xrekson/game-backend/services/economy/model"
	"github.com/xrekson/game-backend/services/economy/repository"
)

type EconomyService interface {
	GetBalance(ctx context.Context, playerID string) (*model.PlayerBalance, float64, error)
	ClaimIncome(ctx context.Context, playerID string) (*model.PlayerBalance, float64, error)
	GetUpgrades(ctx context.Context, playerID string) ([]*model.UpgradeConfig, map[string]int32, error)
	PurchaseUpgrade(ctx context.Context, playerID, upgradeID string) (*model.PlayerBalance, float64, *model.PlayerUpgrade, error)
	Prestige(ctx context.Context, playerID string) (*model.PlayerBalance, error)
	DeductCurrency(ctx context.Context, playerID string, amount float64) (float64, error)
	AddCurrency(ctx context.Context, playerID string, amount float64) (float64, error)
}

type economyService struct {
	repo     repository.EconomyRepository
	producer *kafka.Producer
}

func NewEconomyService(repo repository.EconomyRepository, producer *kafka.Producer) EconomyService {
	return &economyService{
		repo:     repo,
		producer: producer,
	}
}

func (s *economyService) getOrInitBalance(ctx context.Context, playerID string) (*model.PlayerBalance, error) {
	b, err := s.repo.GetBalance(ctx, playerID)
	if err != nil {
		return nil, errors.Internal("failed to fetch balance", err)
	}
	if b == nil {
		b = &model.PlayerBalance{
			PlayerID:           playerID,
			Balance:            10.0,
			BaseIncomePerSec:   1.0,
			PrestigeLevel:      0,
			PrestigeMultiplier: 1.0,
			LastClaimTime:      time.Now(),
			UpdatedAt:          time.Now(),
		}
		if err := s.repo.CreateBalance(ctx, b); err != nil {
			return nil, errors.Internal("failed to initialize balance", err)
		}
	}
	return b, nil
}

func (s *economyService) calculateIncomePerSec(ctx context.Context, playerID string, b *model.PlayerBalance) (float64, error) {
	upgrades, err := s.repo.GetUpgrades(ctx, playerID)
	if err != nil {
		return 0, err
	}

	upgradeMap := make(map[string]int32)
	for _, u := range upgrades {
		upgradeMap[u.UpgradeID] = u.Level
	}

	income := b.BaseIncomePerSec
	for _, cfg := range model.DefaultUpgrades {
		if lvl, ok := upgradeMap[cfg.ID]; ok {
			income += float64(lvl) * cfg.IncomeBonusPerSec
		}
	}

	return income * b.PrestigeMultiplier, nil
}

func (s *economyService) GetBalance(ctx context.Context, playerID string) (*model.PlayerBalance, float64, error) {
	b, err := s.getOrInitBalance(ctx, playerID)
	if err != nil {
		return nil, 0, err
	}

	incomePerSec, err := s.calculateIncomePerSec(ctx, playerID, b)
	if err != nil {
		return nil, 0, errors.Internal("failed calculating income per sec", err)
	}

	return b, incomePerSec, nil
}

func (s *economyService) ClaimIncome(ctx context.Context, playerID string) (*model.PlayerBalance, float64, error) {
	b, err := s.getOrInitBalance(ctx, playerID)
	if err != nil {
		return nil, 0, err
	}

	incomePerSec, err := s.calculateIncomePerSec(ctx, playerID, b)
	if err != nil {
		return nil, 0, errors.Internal("failed calculating income per sec", err)
	}

	now := time.Now()
	secondsElapsed := now.Sub(b.LastClaimTime).Seconds()
	// Cap offline income to 24 hours (86400 seconds)
	if secondsElapsed > 86400 {
		secondsElapsed = 86400
	}

	claimed := secondsElapsed * incomePerSec
	b.Balance += claimed
	b.LastClaimTime = now

	if err := s.repo.UpdateBalance(ctx, b); err != nil {
		return nil, 0, errors.Internal("failed updating claimed balance", err)
	}

	if s.producer != nil {
		_ = s.producer.Publish(ctx, events.TopicEconomyIncome, playerID, events.IncomeClaimedEvent{
			PlayerID:     playerID,
			Amount:       claimed,
			TotalBalance: b.Balance,
			Timestamp:    now,
		})
	}

	return b, claimed, nil
}

func (s *economyService) GetUpgrades(ctx context.Context, playerID string) ([]*model.UpgradeConfig, map[string]int32, error) {
	upgrades, err := s.repo.GetUpgrades(ctx, playerID)
	if err != nil {
		return nil, nil, errors.Internal("failed to fetch player upgrades", err)
	}

	levelMap := make(map[string]int32)
	for _, u := range upgrades {
		levelMap[u.UpgradeID] = u.Level
	}

	configs := make([]*model.UpgradeConfig, len(model.DefaultUpgrades))
	for i, cfg := range model.DefaultUpgrades {
		lvl := levelMap[cfg.ID]
		c := cfg
		c.BaseCost = cfg.BaseCost * math.Pow(1.15, float64(lvl))
		configs[i] = &c
	}

	return configs, levelMap, nil
}

func (s *economyService) PurchaseUpgrade(ctx context.Context, playerID, upgradeID string) (*model.PlayerBalance, float64, *model.PlayerUpgrade, error) {
	b, incomePerSec, err := s.GetBalance(ctx, playerID)
	if err != nil {
		return nil, 0, nil, err
	}

	var targetConfig *model.UpgradeConfig
	for _, cfg := range model.DefaultUpgrades {
		if cfg.ID == upgradeID {
			targetConfig = &cfg
			break
		}
	}
	if targetConfig == nil {
		return nil, 0, nil, errors.NotFound("upgrade not found", nil)
	}

	u, err := s.repo.GetUpgrade(ctx, playerID, upgradeID)
	if err != nil {
		return nil, 0, nil, errors.Internal("failed fetching upgrade level", err)
	}
	if u == nil {
		u = &model.PlayerUpgrade{PlayerID: playerID, UpgradeID: upgradeID, Level: 0}
	}

	cost := targetConfig.BaseCost * math.Pow(1.15, float64(u.Level))
	if b.Balance < cost {
		return nil, 0, nil, errors.BadRequest(fmt.Sprintf("insufficient funds for upgrade: requires %.2f", cost), nil)
	}

	b.Balance -= cost
	u.Level++

	if err := s.repo.UpdateBalance(ctx, b); err != nil {
		return nil, 0, nil, errors.Internal("failed updating balance after upgrade", err)
	}

	if err := s.repo.SaveUpgrade(ctx, u); err != nil {
		return nil, 0, nil, errors.Internal("failed saving upgrade", err)
	}

	newIncomePerSec, err := s.calculateIncomePerSec(ctx, playerID, b)
	if err != nil {
		newIncomePerSec = incomePerSec
	}

	if s.producer != nil {
		_ = s.producer.Publish(ctx, events.TopicEconomyUpgrades, playerID, events.UpgradePurchasedEvent{
			PlayerID:     playerID,
			UpgradeID:    upgradeID,
			Cost:         cost,
			NewLevel:     int(u.Level),
			IncomePerSec: newIncomePerSec,
			Timestamp:    time.Now(),
		})
	}

	return b, newIncomePerSec, u, nil
}

func (s *economyService) Prestige(ctx context.Context, playerID string) (*model.PlayerBalance, error) {
	b, _, err := s.GetBalance(ctx, playerID)
	if err != nil {
		return nil, err
	}

	const prestigeThreshold = 100000.0
	if b.Balance < prestigeThreshold {
		return nil, errors.BadRequest(fmt.Sprintf("prestige requires at least %.0f balance", prestigeThreshold), nil)
	}

	b.PrestigeLevel++
	b.PrestigeMultiplier *= 1.5
	b.Balance = 10.0
	b.LastClaimTime = time.Now()

	if err := s.repo.UpdateBalance(ctx, b); err != nil {
		return nil, errors.Internal("failed updating prestige balance", err)
	}

	return b, nil
}

func (s *economyService) DeductCurrency(ctx context.Context, playerID string, amount float64) (float64, error) {
	newBal, err := s.repo.DeductCurrency(ctx, playerID, amount)
	if err != nil {
		return 0, errors.BadRequest("insufficient balance or database error", err)
	}
	return newBal, nil
}

func (s *economyService) AddCurrency(ctx context.Context, playerID string, amount float64) (float64, error) {
	newBal, err := s.repo.AddCurrency(ctx, playerID, amount)
	if err != nil {
		return 0, errors.Internal("failed adding currency", err)
	}
	return newBal, nil
}
