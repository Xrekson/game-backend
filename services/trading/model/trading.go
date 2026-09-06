package model

import "time"

type TradeListing struct {
	ID        string    `json:"id"`
	SellerID  string    `json:"seller_id"`
	ItemID    string    `json:"item_id"`
	ItemName  string    `json:"item_name"`
	Price     float64   `json:"price"`
	Status    string    `json:"status"` // active, completed, cancelled
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OutboxEvent struct {
	ID        string    `json:"id"`
	Topic     string    `json:"topic"`
	EventKey  string    `json:"event_key"`
	Payload   string    `json:"payload"`
	Processed bool      `json:"processed"`
	CreatedAt time.Time `json:"created_at"`
}
