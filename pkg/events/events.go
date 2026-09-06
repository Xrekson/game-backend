package events

import "time"

const (
	TopicEconomyIncome      = "economy.income"
	TopicEconomyUpgrades    = "economy.upgrades"
	TopicTradingCreated     = "trading.created"
	TopicTradingCompleted   = "trading.completed"
	TopicWorldCombat        = "world.combat"
	TopicWorldPresence      = "world.presence"
)

type IncomeClaimedEvent struct {
	PlayerID     string    `json:"player_id"`
	Amount       float64   `json:"amount"`
	TotalBalance float64   `json:"total_balance"`
	Timestamp    time.Time `json:"timestamp"`
}

type UpgradePurchasedEvent struct {
	PlayerID     string    `json:"player_id"`
	UpgradeID    string    `json:"upgrade_id"`
	Cost         float64   `json:"cost"`
	NewLevel     int       `json:"new_level"`
	IncomePerSec float64   `json:"income_per_sec"`
	Timestamp    time.Time `json:"timestamp"`
}

type TradeCreatedEvent struct {
	TradeID   string    `json:"trade_id"`
	SellerID  string    `json:"seller_id"`
	ItemID    string    `json:"item_id"`
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}

type TradeCompletedEvent struct {
	TradeID   string    `json:"trade_id"`
	BuyerID   string    `json:"buyer_id"`
	SellerID  string    `json:"seller_id"`
	ItemID    string    `json:"item_id"`
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}

type CombatEvent struct {
	PlayerID    string    `json:"player_id"`
	DungeonID   string    `json:"dungeon_id"`
	EnemyID     string    `json:"enemy_id"`
	DamageDealt float64   `json:"damage_dealt"`
	PlayerHP    float64   `json:"player_hp"`
	EnemyHP     float64   `json:"enemy_hp"`
	Result      string    `json:"result"` // "victory", "defeat", "in_progress"
	Timestamp   time.Time `json:"timestamp"`
}

type PresenceEvent struct {
	PlayerID   string    `json:"player_id"`
	MapID      string    `json:"map_id"`
	X          float64   `json:"x"`
	Y          float64   `json:"y"`
	IsSafeZone bool      `json:"is_safe_zone"`
	Timestamp  time.Time `json:"timestamp"`
}
