package model

import "time"

type PlayerProfile struct {
	PlayerID  string    `json:"player_id"`
	Username  string    `json:"username"`
	Level     int32     `json:"level"`
	XP        int64     `json:"xp"`
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
