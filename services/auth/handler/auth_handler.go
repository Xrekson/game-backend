package handler

import (
	"context"

	"github.com/xrekson/game-backend/proto/auth"
	"github.com/xrekson/game-backend/services/auth/service"
)

type AuthHandler struct {
	authpb.UnimplementedAuthServiceServer
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
	user, token, err := h.authService.Register(ctx, req.Username, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	return &authpb.RegisterResponse{
		UserId:   user.ID,
		Username: user.Username,
		Token:    token,
	}, nil
}

func (h *AuthHandler) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	user, token, err := h.authService.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	return &authpb.LoginResponse{
		UserId:   user.ID,
		Username: user.Username,
		Token:    token,
	}, nil
}

func (h *AuthHandler) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
	claims, err := h.authService.ValidateToken(ctx, req.Token)
	if err != nil {
		return &authpb.ValidateTokenResponse{Valid: false}, nil
	}

	return &authpb.ValidateTokenResponse{
		Valid:    true,
		UserId:   claims.UserID,
		Username: claims.Username,
	}, nil
}
