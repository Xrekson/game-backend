package main

import (
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"

	"github.com/xrekson/game-backend/pkg/auth"
	"github.com/xrekson/game-backend/pkg/config"
	"github.com/xrekson/game-backend/pkg/logger"
	"github.com/xrekson/game-backend/pkg/postgres"
	"github.com/xrekson/game-backend/proto/auth"
	"github.com/xrekson/game-backend/services/auth/handler"
	"github.com/xrekson/game-backend/services/auth/repository"
	"github.com/xrekson/game-backend/services/auth/service"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/dev.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("Failed to load config: %v", err)
		os.Exit(1)
	}

	db, err := postgres.NewPostgresDB(cfg.Postgres)
	if err != nil {
		logger.Error("Failed to connect to postgres: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	jwtMgr := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.AccessExpiryHours)
	authSvc := service.NewAuthService(userRepo, jwtMgr)
	authHdlr := handler.NewAuthHandler(authSvc)

	port := cfg.Services.AuthGRPCPort
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Error("Failed to listen on port %d: %v", port, err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	authpb.RegisterAuthServiceServer(grpcServer, authHdlr)

	logger.Info("Auth Service gRPC server running on port %d", port)
	if err := grpcServer.Serve(lis); err != nil {
		logger.Error("gRPC server failed: %v", err)
	}
}
