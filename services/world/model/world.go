package model

type Position struct {
	X float64 `json:"x" bson:"x"`
	Y float64 `json:"y" bson:"y"`
}

type MapLayout struct {
	MapID      string   `json:"map_id" bson:"map_id"`
	Name       string   `json:"name" bson:"name"`
	Width      float64  `json:"width" bson:"width"`
	Height     float64  `json:"height" bson:"height"`
	IsSafeZone bool     `json:"is_safe_zone" bson:"is_safe_zone"`
	SpawnPoint Position `json:"spawn_point" bson:"spawn_point"`
}

type PlayerState struct {
	PlayerID   string   `json:"player_id"`
	MapID      string   `json:"map_id"`
	Position   Position `json:"position"`
	HP         float64  `json:"hp"`
	MaxHP      float64  `json:"max_hp"`
	IsSafeZone bool     `json:"is_safe_zone"`
}

type WSMessage struct {
	Type    string `json:"type"` // "move", "combat", "presence", "chat"
	Payload any    `json:"payload"`
}

type MovePayload struct {
	MapID string  `json:"map_id"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
}

type CombatPayload struct {
	EnemyID string  `json:"enemy_id"`
	Damage  float64 `json:"damage"`
}

var DefaultMaps = []MapLayout{
	{
		MapID:      "safe_zone_1",
		Name:       "Haven Town Square",
		Width:      500.0,
		Height:     500.0,
		IsSafeZone: true,
		SpawnPoint: Position{X: 250.0, Y: 250.0},
	},
	{
		MapID:      "dungeon_1",
		Name:       "Shadow Crypt Level 1",
		Width:      1000.0,
		Height:     1000.0,
		IsSafeZone: false,
		SpawnPoint: Position{X: 50.0, Y: 50.0},
	},
}
