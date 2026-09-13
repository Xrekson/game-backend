package service

import (
	"context"
	"fmt"
	"math"
	"strings"
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
	Prestige(ctx context.Context, playerID string) (*model.PlayerBalance, float64, error)
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

	income := b.BaseIncomePerSec
	for _, u := range upgrades {
		if u.Level <= 0 {
			continue
		}
		baseID, rank := model.ParseUpgradeIDAndRank(u.UpgradeID)
		for _, cfg := range model.DefaultUpgrades {
			if cfg.ID == baseID {
				rankCfg := model.GetConfigForRank(cfg, rank)
				income += float64(u.Level) * rankCfg.IncomeBonusPerSec
				break
			}
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
	b, err := s.getOrInitBalance(ctx, playerID)
	if err != nil {
		return nil, nil, errors.Internal("failed to fetch balance for upgrades", err)
	}

	upgrades, err := s.repo.GetUpgrades(ctx, playerID)
	if err != nil {
		return nil, nil, errors.Internal("failed to fetch player upgrades", err)
	}

	levelMap := make(map[string]int32)
	for _, u := range upgrades {
		levelMap[u.UpgradeID] = u.Level
	}

	currentRank := b.PrestigeLevel
	var configs []*model.UpgradeConfig

	// 1. Active shop upgrades for the current prestige rank
	for _, baseCfg := range model.DefaultUpgrades {
		rankCfg := model.GetConfigForRank(baseCfg, currentRank)
		lvl := levelMap[rankCfg.ID]
		c := rankCfg
		c.BaseCost = rankCfg.BaseCost * math.Pow(1.15, float64(lvl))
		configs = append(configs, &c)
	}

	// 2. Historical upgrades from previous ranks that have level > 0
	activeMap := make(map[string]bool)
	for _, c := range configs {
		activeMap[c.ID] = true
	}

	for upgradeID, lvl := range levelMap {
		if lvl <= 0 || activeMap[upgradeID] {
			continue
		}
		baseID, rank := model.ParseUpgradeIDAndRank(upgradeID)
		for _, baseCfg := range model.DefaultUpgrades {
			if baseCfg.ID == baseID {
				rankCfg := model.GetConfigForRank(baseCfg, rank)
				c := rankCfg
				c.BaseCost = rankCfg.BaseCost * math.Pow(1.15, float64(lvl))
				configs = append(configs, &c)
				break
			}
		}
	}

	return configs, levelMap, nil
}

func (s *economyService) PurchaseUpgrade(ctx context.Context, playerID, upgradeID string) (*model.PlayerBalance, float64, *model.PlayerUpgrade, error) {
	b, incomePerSec, err := s.GetBalance(ctx, playerID)
	if err != nil {
		return nil, 0, nil, err
	}

	targetUpgradeID := upgradeID
	baseID, rank := model.ParseUpgradeIDAndRank(upgradeID)

	var targetConfig *model.UpgradeConfig
	for _, cfg := range model.DefaultUpgrades {
		if cfg.ID == baseID {
			resolvedRank := rank
			if !strings.Contains(upgradeID, "_r") {
				resolvedRank = b.PrestigeLevel
				targetUpgradeID = fmt.Sprintf("%s_r%d", baseID, resolvedRank)
			}
			rankCfg := model.GetConfigForRank(cfg, resolvedRank)
			targetConfig = &rankCfg
			break
		}
	}

	if targetConfig == nil {
		return nil, 0, nil, errors.NotFound("upgrade not found", nil)
	}

	u, err := s.repo.GetUpgrade(ctx, playerID, targetUpgradeID)
	if err != nil {
		return nil, 0, nil, errors.Internal("failed fetching upgrade level", err)
	}
	if u == nil {
		u = &model.PlayerUpgrade{PlayerID: playerID, UpgradeID: targetUpgradeID, Level: 0}
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
			UpgradeID:    targetUpgradeID,
			Cost:         cost,
			NewLevel:     int(u.Level),
			IncomePerSec: newIncomePerSec,
			Timestamp:    time.Now(),
		})
	}

	return b, newIncomePerSec, u, nil
}

func (s *economyService) Prestige(ctx context.Context, playerID string) (*model.PlayerBalance, float64, error) {
	b, _, err := s.GetBalance(ctx, playerID)
	if err != nil {
		return nil, 0, err
	}

	incomePerSec, err := s.calculateIncomePerSec(ctx, playerID, b)
	if err != nil {
		return nil, 0, errors.Internal("failed calculating income for prestige", err)
	}

	// Target income rate required for current rank: 5000 * (1.5 ^ PrestigeLevel)
	targetIncomeRate := 5000.0 * math.Pow(1.5, float64(b.PrestigeLevel))
	if incomePerSec < targetIncomeRate {
		return nil, 0, errors.BadRequest(fmt.Sprintf("prestige rank up requires at least %.0f Gold/sec income rate (current: %.1f Gold/sec)", targetIncomeRate, incomePerSec), nil)
	}

	b.PrestigeLevel++
	// Multiplier formula: 1.0 + 0.5 * PrestigeLevel (Rank 0 = 1.0, Rank 1 = 1.5, Rank 2 = 2.0...)
	b.PrestigeMultiplier = 1.0 + 0.5*float64(b.PrestigeLevel)
	b.Balance = 10.0
	b.LastClaimTime = time.Now()

	if err := s.repo.UpdateBalance(ctx, b); err != nil {
		return nil, 0, errors.Internal("failed updating prestige balance", err)
	}

	newIncomePerSec, _ := s.calculateIncomePerSec(ctx, playerID, b)

	return b, newIncomePerSec, nil
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
