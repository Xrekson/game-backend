package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/xrekson/game-backend/pkg/config"
	"github.com/xrekson/game-backend/pkg/logger"
)

type DB struct {
	*sql.DB
}

func NewPostgresDB(cfg config.PostgresConfig) (*DB, error) {
	dsn := cfg.DSN()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		logger.Warn("PostgreSQL ping failed, initial database connection pending: %v", err)
	} else {
		logger.Info("Connected to PostgreSQL successfully at %s:%d", cfg.Host, cfg.Port)
	}

	return &DB{DB: db}, nil
}
