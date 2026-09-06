package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/xrekson/game-backend/pkg/postgres"
	"github.com/xrekson/game-backend/services/trading/model"
)

type TradingRepository interface {
	CreateListing(ctx context.Context, listing *model.TradeListing) error
	GetListings(ctx context.Context, status string) ([]*model.TradeListing, error)
	GetListingByID(ctx context.Context, id string) (*model.TradeListing, error)
	UpdateListingStatus(ctx context.Context, id, status string) error
	SaveOutboxEvent(ctx context.Context, event *model.OutboxEvent) error
	GetPendingOutboxEvents(ctx context.Context, limit int) ([]*model.OutboxEvent, error)
	MarkOutboxEventProcessed(ctx context.Context, id string) error
}

type postgresTradingRepository struct {
	db *postgres.DB
}

func NewTradingRepository(db *postgres.DB) TradingRepository {
	return &postgresTradingRepository{db: db}
}

func (r *postgresTradingRepository) CreateListing(ctx context.Context, l *model.TradeListing) error {
	query := `INSERT INTO trade_listings (id, seller_id, item_id, price, status, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query, l.ID, l.SellerID, l.ItemID, l.Price, l.Status, l.CreatedAt, l.UpdatedAt)
	return err
}

func (r *postgresTradingRepository) GetListings(ctx context.Context, status string) ([]*model.TradeListing, error) {
	var query string
	var args []any

	if status != "" {
		query = `SELECT id, seller_id, item_id, price, status, created_at, updated_at FROM trade_listings WHERE status = $1 ORDER BY created_at DESC`
		args = append(args, status)
	} else {
		query = `SELECT id, seller_id, item_id, price, status, created_at, updated_at FROM trade_listings ORDER BY created_at DESC`
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var listings []*model.TradeListing
	for rows.Next() {
		l := &model.TradeListing{}
		if err := rows.Scan(&l.ID, &l.SellerID, &l.ItemID, &l.Price, &l.Status, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, err
		}
		listings = append(listings, l)
	}
	return listings, nil
}

func (r *postgresTradingRepository) GetListingByID(ctx context.Context, id string) (*model.TradeListing, error) {
	query := `SELECT id, seller_id, item_id, price, status, created_at, updated_at FROM trade_listings WHERE id = $1`
	l := &model.TradeListing{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(&l.ID, &l.SellerID, &l.ItemID, &l.Price, &l.Status, &l.CreatedAt, &l.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return l, nil
}

func (r *postgresTradingRepository) UpdateListingStatus(ctx context.Context, id, status string) error {
	query := `UPDATE trade_listings SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	return err
}

func (r *postgresTradingRepository) SaveOutboxEvent(ctx context.Context, e *model.OutboxEvent) error {
	query := `INSERT INTO outbox_events (id, topic, event_key, payload, processed, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query, e.ID, e.Topic, e.EventKey, e.Payload, e.Processed, e.CreatedAt)
	return err
}

func (r *postgresTradingRepository) GetPendingOutboxEvents(ctx context.Context, limit int) ([]*model.OutboxEvent, error) {
	query := `SELECT id, topic, event_key, payload, processed, created_at 
			  FROM outbox_events WHERE processed = FALSE ORDER BY created_at ASC LIMIT $1`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*model.OutboxEvent
	for rows.Next() {
		e := &model.OutboxEvent{}
		if err := rows.Scan(&e.ID, &e.Topic, &e.EventKey, &e.Payload, &e.Processed, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func (r *postgresTradingRepository) MarkOutboxEventProcessed(ctx context.Context, id string) error {
	query := `UPDATE outbox_events SET processed = TRUE WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
