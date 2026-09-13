package model

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

type PlayerBalance struct {
	PlayerID           string    `json:"player_id"`
	Balance            float64   `json:"balance"`
	BaseIncomePerSec   float64   `json:"base_income_per_sec"`
	PrestigeLevel      int32     `json:"prestige_level"`
	PrestigeMultiplier float64   `json:"prestige_multiplier"`
	LastClaimTime      time.Time `json:"last_claim_time"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type PlayerUpgrade struct {
	PlayerID  string    `json:"player_id"`
	UpgradeID string    `json:"upgrade_id"`
	Level     int32     `json:"level"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpgradeConfig struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	BaseCost          float64 `json:"base_cost"`
	IncomeBonusPerSec float64 `json:"income_bonus_per_sec"`
	Description       string  `json:"description"`
}

var DefaultUpgrades = []UpgradeConfig{
	{ID: "miner_1", Name: "Gold Miner", BaseCost: 15.0, IncomeBonusPerSec: 0.5, Description: "Basic miner digging for gold."},
	{ID: "forge_1", Name: "Blacksmith Forge", BaseCost: 100.0, IncomeBonusPerSec: 4.0, Description: "Forges weapons for additional revenue."},
	{ID: "dungeon_extract", Name: "Mana Crystal Extractor", BaseCost: 1100.0, IncomeBonusPerSec: 32.0, Description: "Extracts mana crystals from deep dungeons."},
	{ID: "trade_guild", Name: "Merchant Guild Investment", BaseCost: 12000.0, IncomeBonusPerSec: 260.0, Description: "Automated trade revenues."},
}

func ParseUpgradeIDAndRank(upgradeID string) (string, int32) {
	idx := strings.LastIndex(upgradeID, "_r")
	if idx != -1 {
		base := upgradeID[:idx]
		rankStr := upgradeID[idx+2:]
		if r, err := strconv.Atoi(rankStr); err == nil {
			return base, int32(r)
		}
	}
	return upgradeID, 0
}

func GetConfigForRank(baseCfg UpgradeConfig, rank int32) UpgradeConfig {
	rankMultiplier := math.Pow(2.5, float64(rank))
	incomeMultiplier := math.Pow(2.0, float64(rank))
	return UpgradeConfig{
		ID:                fmt.Sprintf("%s_r%d", baseCfg.ID, rank),
		Name:              fmt.Sprintf("%s (Rank %d)", baseCfg.Name, rank),
		BaseCost:          baseCfg.BaseCost * rankMultiplier,
		IncomeBonusPerSec: baseCfg.IncomeBonusPerSec * incomeMultiplier,
		Description:       fmt.Sprintf("%s [Rank %d Tier]", baseCfg.Description, rank),
	}
}

