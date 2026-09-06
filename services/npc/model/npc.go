package model

type NPC struct {
	ID                   string   `json:"id" bson:"id"`
	Name                 string   `json:"name" bson:"name"`
	MapID                string   `json:"map_id" bson:"map_id"`
	Dialogue             string   `json:"dialogue" bson:"dialogue"`
	ShopItemTemplateIDs []string `json:"shop_item_template_ids" bson:"shop_item_template_ids"`
}

var DefaultNPCs = []NPC{
	{
		ID:                   "npc_blacksmith",
		Name:                 "Garrick the Blacksmith",
		MapID:                "safe_zone_1",
		Dialogue:             "Welcome traveler! Look at my weapons and armours forged in the safe haven.",
		ShopItemTemplateIDs: []string{"wep_iron_sword", "wep_mana_blade", "arm_leather_vest", "arm_plate_armour"},
	},
	{
		ID:                   "npc_alchemist",
		Name:                 "Lyra the Mystic",
		MapID:                "safe_zone_1",
		Dialogue:             "Seeking grand power? Behold the legendary artifacts.",
		ShopItemTemplateIDs: []string{"arm_ring_power"},
	},
}
