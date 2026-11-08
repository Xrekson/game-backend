package main

import (
	"fmt"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/xrekson/game-backend/gateway/router"
	"github.com/xrekson/game-backend/pkg/auth"
	"github.com/xrekson/game-backend/pkg/config"
	"github.com/xrekson/game-backend/pkg/logger"
	"github.com/xrekson/game-backend/pkg/redis"
	"github.com/xrekson/game-backend/proto/auth"
	"github.com/xrekson/game-backend/proto/economy"
	"github.com/xrekson/game-backend/proto/inventory"
	"github.com/xrekson/game-backend/proto/npc"
	"github.com/xrekson/game-backend/proto/player"
	"github.com/xrekson/game-backend/proto/trading"
	"github.com/xrekson/game-backend/proto/world"
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

	jwtMgr := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.AccessExpiryHours)

	redisClient, err := redis.NewRedisClient(cfg.Redis)
	if err != nil {
		logger.Error("Failed to connect to redis: %v", err)
	}

	// Dial gRPC clients for all microservices
	dialOptions := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	authConn, _ := grpc.NewClient(fmt.Sprintf("localhost:%d", cfg.Services.AuthGRPCPort), dialOptions...)
	defer authConn.Close()
	authClient := authpb.NewAuthServiceClient(authConn)

	playerConn, _ := grpc.NewClient(fmt.Sprintf("localhost:%d", cfg.Services.PlayerGRPCPort), dialOptions...)
	defer playerConn.Close()
	playerClient := playerpb.NewPlayerServiceClient(playerConn)

	econConn, _ := grpc.NewClient(fmt.Sprintf("localhost:%d", cfg.Services.EconomyGRPCPort), dialOptions...)
	defer econConn.Close()
	econClient := economypb.NewEconomyServiceClient(econConn)

	invConn, _ := grpc.NewClient(fmt.Sprintf("localhost:%d", cfg.Services.InventoryGRPCPort), dialOptions...)
	defer invConn.Close()
	invClient := inventorypb.NewInventoryServiceClient(invConn)

	tradingConn, _ := grpc.NewClient(fmt.Sprintf("localhost:%d", cfg.Services.TradingGRPCPort), dialOptions...)
	defer tradingConn.Close()
	tradingClient := tradingpb.NewTradingServiceClient(tradingConn)

	npcConn, _ := grpc.NewClient(fmt.Sprintf("localhost:%d", cfg.Services.NPCGRPCPort), dialOptions...)
	defer npcConn.Close()
	npcClient := npcpb.NewNPCServiceClient(npcConn)

	worldConn, _ := grpc.NewClient(fmt.Sprintf("localhost:%d", cfg.Services.WorldGRPCPort), dialOptions...)
	defer worldConn.Close()
	worldClient := worldpb.NewWorldServiceClient(worldConn)

	clients := router.Clients{
		Auth:      authClient,
		Player:    playerClient,
		Economy:   econClient,
		Inventory: invClient,
		Trading:   tradingClient,
		NPC:       npcClient,
		World:     worldClient,
	}

	rateLimitRPM := cfg.Services.RateLimitRPM
	if rateLimitRPM == 0 {
		rateLimitRPM = 120
	}
	r := router.SetupRouter(jwtMgr, redisClient, clients, rateLimitRPM)

	port := cfg.Services.GatewayPort
	logger.Info("API Gateway running on port %d", port)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		logger.Error("API Gateway failed: %v", err)
	}
}
