package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/xrekson/game-backend/pkg/postgres"
	"github.com/xrekson/game-backend/services/auth/model"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
	FindByUsername(ctx context.Context, username string) (*model.User, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
}

type postgresUserRepository struct {
	db *postgres.DB
}

func NewUserRepository(db *postgres.DB) UserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) CreateUser(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (id, username, email, password_hash, role, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Username, user.Email, user.PasswordHash, user.Role, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}
	return nil
}

func (r *postgresUserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `SELECT id, username, email, password_hash, role, created_at, updated_at FROM users WHERE username = $1`
	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *postgresUserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	query := `SELECT id, username, email, password_hash, role, created_at, updated_at FROM users WHERE id = $1`
	user := &model.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}
