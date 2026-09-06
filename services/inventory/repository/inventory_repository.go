package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/xrekson/game-backend/pkg/postgres"
	"github.com/xrekson/game-backend/services/inventory/model"
)

type InventoryRepository interface {
	GetInventory(ctx context.Context, playerID string) ([]*model.InventoryItem, error)
	GetItemByID(ctx context.Context, itemID string) (*model.InventoryItem, error)
	AddItem(ctx context.Context, item *model.InventoryItem) error
	RemoveItem(ctx context.Context, itemID string) error
	SetEquipped(ctx context.Context, itemID string, equipped bool) error
	UnequipAllOfType(ctx context.Context, playerID string, itemType model.ItemType) error
}

type postgresInventoryRepository struct {
	db *postgres.DB
}

func NewInventoryRepository(db *postgres.DB) InventoryRepository {
	return &postgresInventoryRepository{db: db}
}

func (r *postgresInventoryRepository) GetInventory(ctx context.Context, playerID string) ([]*model.InventoryItem, error) {
	query := `SELECT id, player_id, template_id, name, item_type, attack, defense, rarity, is_equipped, price, has_attributes, attributes, created_at 
			  FROM player_inventory WHERE player_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*model.InventoryItem
	for rows.Next() {
		item := &model.InventoryItem{}
		var attrsJSON []byte
		if err := rows.Scan(
			&item.ID, &item.PlayerID, &item.TemplateID, &item.Name, &item.Type,
			&item.Attack, &item.Defense, &item.Rarity, &item.IsEquipped, &item.Price,
			&item.HasAttributes, &attrsJSON, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		if len(attrsJSON) > 0 {
			_ = json.Unmarshal(attrsJSON, &item.Attributes)
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *postgresInventoryRepository) GetItemByID(ctx context.Context, itemID string) (*model.InventoryItem, error) {
	query := `SELECT id, player_id, template_id, name, item_type, attack, defense, rarity, is_equipped, price, has_attributes, attributes, created_at 
			  FROM player_inventory WHERE id = $1`
	item := &model.InventoryItem{}
	var attrsJSON []byte
	err := r.db.QueryRowContext(ctx, query, itemID).Scan(
		&item.ID, &item.PlayerID, &item.TemplateID, &item.Name, &item.Type,
		&item.Attack, &item.Defense, &item.Rarity, &item.IsEquipped, &item.Price,
		&item.HasAttributes, &attrsJSON, &item.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if len(attrsJSON) > 0 {
		_ = json.Unmarshal(attrsJSON, &item.Attributes)
	}
	return item, nil
}

func (r *postgresInventoryRepository) AddItem(ctx context.Context, item *model.InventoryItem) error {
	attrsJSON, _ := json.Marshal(item.Attributes)
	query := `INSERT INTO player_inventory (id, player_id, template_id, name, item_type, attack, defense, rarity, is_equipped, price, has_attributes, attributes, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	_, err := r.db.ExecContext(ctx, query,
		item.ID, item.PlayerID, item.TemplateID, item.Name, item.Type,
		item.Attack, item.Defense, item.Rarity, item.IsEquipped, item.Price,
		item.HasAttributes, attrsJSON, item.CreatedAt,
	)
	return err
}

func (r *postgresInventoryRepository) RemoveItem(ctx context.Context, itemID string) error {
	query := `DELETE FROM player_inventory WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, itemID)
	return err
}

func (r *postgresInventoryRepository) SetEquipped(ctx context.Context, itemID string, equipped bool) error {
	query := `UPDATE player_inventory SET is_equipped = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, equipped, itemID)
	return err
}

func (r *postgresInventoryRepository) UnequipAllOfType(ctx context.Context, playerID string, itemType model.ItemType) error {
	query := `UPDATE player_inventory SET is_equipped = FALSE WHERE player_id = $1 AND item_type = $2`
	_, err := r.db.ExecContext(ctx, query, playerID, itemType)
	return err
}
