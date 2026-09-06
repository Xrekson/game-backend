package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/xrekson/game-backend/pkg/auth"
	appErrors "github.com/xrekson/game-backend/pkg/errors"
	"github.com/xrekson/game-backend/services/auth/model"
	"github.com/xrekson/game-backend/services/auth/repository"
)

type AuthService interface {
	Register(ctx context.Context, username, email, password string) (*model.User, string, error)
	Login(ctx context.Context, username, password string) (*model.User, string, error)
	ValidateToken(ctx context.Context, token string) (*auth.Claims, error)
}

type authService struct {
	userRepo repository.UserRepository
	jwtMgr   *auth.JWTManager
}

func NewAuthService(userRepo repository.UserRepository, jwtMgr *auth.JWTManager) AuthService {
	return &authService{
		userRepo: userRepo,
		jwtMgr:   jwtMgr,
	}
}

func (s *authService) Register(ctx context.Context, username, email, password string) (*model.User, string, error) {
	existing, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, "", appErrors.Internal("failed checking username", err)
	}
	if existing != nil {
		return nil, "", appErrors.Conflict("username already taken", nil)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", appErrors.Internal("failed to hash password", err)
	}

	user := &model.User{
		ID:           uuid.New().String(),
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         "player",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, "", appErrors.Internal("failed creating user", err)
	}

	token, err := s.jwtMgr.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, "", appErrors.Internal("failed generating token", err)
	}

	return user, token, nil
}

func (s *authService) Login(ctx context.Context, username, password string) (*model.User, string, error) {
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, "", appErrors.Internal("db lookup error", err)
	}
	if user == nil {
		return nil, "", appErrors.Unauthorized("invalid credentials", nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", appErrors.Unauthorized("invalid credentials", nil)
	}

	token, err := s.jwtMgr.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, "", appErrors.Internal("token generation error", err)
	}

	return user, token, nil
}

func (s *authService) ValidateToken(ctx context.Context, tokenStr string) (*auth.Claims, error) {
	claims, err := s.jwtMgr.ValidateToken(tokenStr)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}
	return claims, nil
}
