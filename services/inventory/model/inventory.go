package model

import "time"

type ItemType int32

const (
	ItemTypeWeapon   ItemType = 0
	ItemTypeArmour   ItemType = 1
	ItemTypeArmament ItemType = 2
)

type InventoryItem struct {
	ID         string    `json:"id"`
	PlayerID   string    `json:"player_id"`
	TemplateID string    `json:"template_id"`
	Name       string    `json:"name"`
	Type       ItemType  `json:"type"`
	Attack     float64   `json:"attack"`
	Defense    float64   `json:"defense"`
	Rarity     string    `json:"rarity"`
	IsEquipped bool      `json:"is_equipped"`
	Price      float64   `json:"price"`
	CreatedAt  time.Time `json:"created_at"`
}

type ItemTemplate struct {
	TemplateID  string   `json:"template_id" bson:"template_id"`
	Name        string   `json:"name" bson:"name"`
	Type        ItemType `json:"type" bson:"type"`
	BaseAttack  float64  `json:"base_attack" bson:"base_attack"`
	BaseDefense float64  `json:"base_defense" bson:"base_defense"`
	Rarity      string   `json:"rarity" bson:"rarity"`
	BasePrice   float64  `json:"base_price" bson:"base_price"`
}

var DefaultItemTemplates = []ItemTemplate{
	{TemplateID: "wep_iron_sword", Name: "Iron Sword", Type: ItemTypeWeapon, BaseAttack: 15, BaseDefense: 0, Rarity: "Common", BasePrice: 50.0},
	{TemplateID: "wep_mana_blade", Name: "Mana Blade", Type: ItemTypeWeapon, BaseAttack: 50, BaseDefense: 5, Rarity: "Rare", BasePrice: 350.0},
	{TemplateID: "arm_leather_vest", Name: "Leather Vest", Type: ItemTypeArmour, BaseAttack: 0, BaseDefense: 10, Rarity: "Common", BasePrice: 40.0},
	{TemplateID: "arm_plate_armour", Name: "Dragon Scale Plate", Type: ItemTypeArmour, BaseAttack: 5, BaseDefense: 60, Rarity: "Epic", BasePrice: 1200.0},
	{TemplateID: "arm_ring_power", Name: "Ring of Power", Type: ItemTypeArmament, BaseAttack: 25, BaseDefense: 25, Rarity: "Legendary", BasePrice: 5000.0},
}
