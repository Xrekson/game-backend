package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/xrekson/game-backend/pkg/config"
	"github.com/xrekson/game-backend/pkg/kafka"
	"github.com/xrekson/game-backend/pkg/logger"
	"github.com/xrekson/game-backend/pkg/postgres"
	"github.com/xrekson/game-backend/proto/economy"
	"github.com/xrekson/game-backend/proto/inventory"
	"github.com/xrekson/game-backend/proto/trading"
	"github.com/xrekson/game-backend/services/trading/handler"
	"github.com/xrekson/game-backend/services/trading/outbox"
	"github.com/xrekson/game-backend/services/trading/repository"
	"github.com/xrekson/game-backend/services/trading/service"
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

	// Connect gRPC clients to Economy and Inventory services
	econConn, err := grpc.NewClient(fmt.Sprintf("localhost:%d", cfg.Services.EconomyGRPCPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("Failed to connect to Economy Service: %v", err)
	}
	defer econConn.Close()
	econClient := economypb.NewEconomyServiceClient(econConn)

	invConn, err := grpc.NewClient(fmt.Sprintf("localhost:%d", cfg.Services.InventoryGRPCPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("Failed to connect to Inventory Service: %v", err)
	}
	defer invConn.Close()
	invClient := inventorypb.NewInventoryServiceClient(invConn)

	tradingRepo := repository.NewTradingRepository(db)

	// Start Outbox Relay worker
	relay := outbox.NewOutboxRelay(tradingRepo, producer)
	relay.Start(2 * time.Second)
	defer relay.Stop()

	tradingSvc := service.NewTradingService(tradingRepo, econClient, invClient)
	tradingHdlr := handler.NewTradingHandler(tradingSvc)

	port := cfg.Services.TradingGRPCPort
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Error("Failed to listen on port %d: %v", port, err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	tradingpb.RegisterTradingServiceServer(grpcServer, tradingHdlr)

	logger.Info("Trading Service gRPC server running on port %d", port)
	if err := grpcServer.Serve(lis); err != nil {
		logger.Error("gRPC server failed: %v", err)
	}
}
