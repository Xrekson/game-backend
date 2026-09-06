package model

import "time"

type ItemType int32

const (
	ItemTypeWeapon   ItemType = 0
	ItemTypeArmour   ItemType = 1
	ItemTypeArmament ItemType = 2
)

type ItemAttribute struct {
	Key             string `json:"key"`
	Type            string `json:"type"`             // "DAMAGE", "BUFF", "EXP_BONUS", "GOLD_BONUS", "SPECIAL"
	Value           string `json:"value"`            // e.g. "15", "1.2", "20%"
	CalculationExpr string `json:"calculation_expr"` // e.g. "base_atk * 1.25", "xp * 1.20"
}

type InventoryItem struct {
	ID            string          `json:"id"`
	PlayerID      string          `json:"player_id"`
	TemplateID    string          `json:"template_id"`
	Name          string          `json:"name"`
	Type          ItemType        `json:"type"`
	Attack        float64         `json:"attack"`
	Defense       float64         `json:"defense"`
	Rarity        string          `json:"rarity"`
	IsEquipped    bool            `json:"is_equipped"`
	Price         float64         `json:"price"`
	HasAttributes bool            `json:"has_attributes"`
	Attributes    []ItemAttribute `json:"attributes"`
	CreatedAt     time.Time       `json:"created_at"`
}

type ItemTemplate struct {
	TemplateID    string          `json:"template_id" bson:"template_id"`
	Name          string          `json:"name" bson:"name"`
	Type          ItemType        `json:"type" bson:"type"`
	BaseAttack    float64         `json:"base_attack" bson:"base_attack"`
	BaseDefense   float64         `json:"base_defense" bson:"base_defense"`
	Rarity        string          `json:"rarity" bson:"rarity"`
	BasePrice     float64         `json:"base_price" bson:"base_price"`
	HasAttributes bool            `json:"has_attributes" bson:"has_attributes"`
	Attributes    []ItemAttribute `json:"attributes" bson:"attributes"`
}

var DefaultItemTemplates = []ItemTemplate{
	{
		TemplateID:    "wep_stone_picker",
		Name:          "Stone Picker",
		Type:          ItemTypeWeapon,
		BaseAttack:    5,
		BaseDefense:   0,
		Rarity:        "Common",
		BasePrice:     25.0,
		HasAttributes: true,
		Attributes: []ItemAttribute{
			{Key: "speed", Type: "BUFF", Value: "1.2", CalculationExpr: "speed * 1.2"},
			{Key: "type", Type: "SPECIAL", Value: "Short Sword / Dagger", CalculationExpr: ""},
		},
	},
	{
		TemplateID:    "wep_iron_gladiator",
		Name:          "Iron Gladiator",
		Type:          ItemTypeWeapon,
		BaseAttack:    15,
		BaseDefense:   0,
		Rarity:        "Rare",
		BasePrice:     250.0,
		HasAttributes: true,
		Attributes: []ItemAttribute{
			{Key: "speed", Type: "BUFF", Value: "1.0", CalculationExpr: "speed * 1.0"},
			{Key: "aeras_crystal", Type: "DAMAGE", Value: "Wind (+4 Physical Atk, 2 Pierce)", CalculationExpr: "base_atk + 4"},
			{Key: "fotia_crystal", Type: "DAMAGE", Value: "Fire (Burn 5/sec for 2s)", CalculationExpr: "burn_dmg = 5 * 2s"},
			{Key: "vronti_crystal", Type: "DAMAGE", Value: "Lightning (Shock 2.5/sec, 20% Stagger)", CalculationExpr: "shock_dmg = 2.5 * 2s"},
		},
	},
	{
		TemplateID:    "wep_havoc_axe",
		Name:          "Havoc Axe",
		Type:          ItemTypeWeapon,
		BaseAttack:    25,
		BaseDefense:   0,
		Rarity:        "Epic",
		BasePrice:     600.0,
		HasAttributes: true,
		Attributes: []ItemAttribute{
			{Key: "speed", Type: "BUFF", Value: "0.75", CalculationExpr: "speed * 0.75"},
			{Key: "element_socket", Type: "SPECIAL", Value: "Middle Crystal Socket", CalculationExpr: "base_atk * 1.25"},
		},
	},
	{
		TemplateID:    "wep_mana_blade",
		Name:          "Mana Blade",
		Type:          ItemTypeWeapon,
		BaseAttack:    50,
		BaseDefense:   5,
		Rarity:        "Rare",
		BasePrice:     350.0,
		HasAttributes: true,
		Attributes: []ItemAttribute{
			{Key: "exp_boost", Type: "EXP_BONUS", Value: "+20% XP", CalculationExpr: "xp * 1.20"},
		},
	},
	{
		TemplateID:    "arm_leather_vest",
		Name:          "Leather Vest",
		Type:          ItemTypeArmour,
		BaseAttack:    0,
		BaseDefense:   10,
		Rarity:        "Common",
		BasePrice:     40.0,
		HasAttributes: false,
		Attributes:    []ItemAttribute{},
	},
	{
		TemplateID:    "arm_plate_armour",
		Name:          "Dragon Scale Plate",
		Type:          ItemTypeArmour,
		BaseAttack:    5,
		BaseDefense:   60,
		Rarity:        "Epic",
		BasePrice:     1200.0,
		HasAttributes: true,
		Attributes: []ItemAttribute{
			{Key: "gold_boost", Type: "GOLD_BONUS", Value: "+15% Income", CalculationExpr: "income * 1.15"},
		},
	},
	{
		TemplateID:    "arm_ring_power",
		Name:          "Ring of Power",
		Type:          ItemTypeArmament,
		BaseAttack:    25,
		BaseDefense:   25,
		Rarity:        "Legendary",
		BasePrice:     5000.0,
		HasAttributes: true,
		Attributes: []ItemAttribute{
			{Key: "power_buff", Type: "BUFF", Value: "All Stats +25%", CalculationExpr: "stats * 1.25"},
		},
	},
}
