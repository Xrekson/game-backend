package main

import (
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"

	"github.com/xrekson/game-backend/pkg/config"
	"github.com/xrekson/game-backend/pkg/logger"
	"github.com/xrekson/game-backend/pkg/postgres"
	"github.com/xrekson/game-backend/proto/player"
	"github.com/xrekson/game-backend/services/player/handler"
	"github.com/xrekson/game-backend/services/player/repository"
	"github.com/xrekson/game-backend/services/player/service"
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

	playerRepo := repository.NewPlayerRepository(db)
	playerSvc := service.NewPlayerService(playerRepo)
	playerHdlr := handler.NewPlayerHandler(playerSvc)

	port := cfg.Services.PlayerGRPCPort
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Error("Failed to listen on port %d: %v", port, err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	playerpb.RegisterPlayerServiceServer(grpcServer, playerHdlr)

	logger.Info("Player Service gRPC server running on port %d", port)
	if err := grpcServer.Serve(lis); err != nil {
		logger.Error("gRPC server failed: %v", err)
	}
}
