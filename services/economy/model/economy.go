package model

import "time"

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
