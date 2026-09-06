package main

import (
	"fmt"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/xrekson/game-backend/pkg/cassandra"
	"github.com/xrekson/game-backend/pkg/config"
	"github.com/xrekson/game-backend/pkg/logger"
	"github.com/xrekson/game-backend/proto/economy"
	"github.com/xrekson/game-backend/proto/inventory"
	"github.com/xrekson/game-backend/proto/npc"
	"github.com/xrekson/game-backend/services/npc/handler"
	"github.com/xrekson/game-backend/services/npc/repository"
	"github.com/xrekson/game-backend/services/npc/service"
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

	cassandraSession, err := cassandra.NewCassandraSession(cfg.Cassandra)
	if err != nil {
		logger.Warn("Cassandra connection notice: %v", err)
	}

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

	npcRepo := repository.NewNPCRepository(cassandraSession)
	npcSvc := service.NewNPCService(npcRepo, econClient, invClient)
	npcHdlr := handler.NewNPCHandler(npcSvc)

	port := cfg.Services.NPCGRPCPort
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Error("Failed to listen on port %d: %v", port, err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	npcpb.RegisterNPCServiceServer(grpcServer, npcHdlr)

	logger.Info("NPC Service gRPC server running on port %d", port)
	if err := grpcServer.Serve(lis); err != nil {
		logger.Error("gRPC server failed: %v", err)
	}
}
