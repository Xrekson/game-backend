package main

import (
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"

	"github.com/xrekson/game-backend/pkg/cassandra"
	"github.com/xrekson/game-backend/pkg/config"
	"github.com/xrekson/game-backend/pkg/logger"
	"github.com/xrekson/game-backend/pkg/postgres"
	"github.com/xrekson/game-backend/proto/inventory"
	"github.com/xrekson/game-backend/services/inventory/handler"
	"github.com/xrekson/game-backend/services/inventory/repository"
	"github.com/xrekson/game-backend/services/inventory/service"
	"github.com/xrekson/game-backend/services/inventory/templaterepo"
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

	cassandraSession, err := cassandra.NewCassandraSession(cfg.Cassandra)
	if err != nil {
		logger.Warn("Cassandra connection notice: %v", err)
	}

	invRepo := repository.NewInventoryRepository(db)
	templateRepo := templaterepo.NewTemplateRepository(cassandraSession)
	invSvc := service.NewInventoryService(invRepo, templateRepo)
	invHdlr := handler.NewInventoryHandler(invSvc)

	port := cfg.Services.InventoryGRPCPort
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Error("Failed to listen on port %d: %v", port, err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	inventorypb.RegisterInventoryServiceServer(grpcServer, invHdlr)

	logger.Info("Inventory Service gRPC server running on port %d", port)
	if err := grpcServer.Serve(lis); err != nil {
		logger.Error("gRPC server failed: %v", err)
	}
}
