package main

import (
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"

	"github.com/xrekson/game-backend/pkg/config"
	"github.com/xrekson/game-backend/pkg/kafka"
	"github.com/xrekson/game-backend/pkg/logger"
	"github.com/xrekson/game-backend/pkg/postgres"
	"github.com/xrekson/game-backend/proto/economy"
	"github.com/xrekson/game-backend/services/economy/handler"
	"github.com/xrekson/game-backend/services/economy/repository"
	"github.com/xrekson/game-backend/services/economy/service"
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

	producer := kafka.NewProducer(cfg.Kafka)
	defer producer.Close()

	econRepo := repository.NewEconomyRepository(db)
	econSvc := service.NewEconomyService(econRepo, producer)
	econHdlr := handler.NewEconomyHandler(econSvc)

	port := cfg.Services.EconomyGRPCPort
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Error("Failed to listen on port %d: %v", port, err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	economypb.RegisterEconomyServiceServer(grpcServer, econHdlr)

	logger.Info("Economy Service gRPC server running on port %d", port)
	if err := grpcServer.Serve(lis); err != nil {
		logger.Error("gRPC server failed: %v", err)
	}
}
