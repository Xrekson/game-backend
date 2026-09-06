package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"

	"google.golang.org/grpc"

	"github.com/xrekson/game-backend/pkg/cassandra"
	"github.com/xrekson/game-backend/pkg/config"
	"github.com/xrekson/game-backend/pkg/kafka"
	"github.com/xrekson/game-backend/pkg/logger"
	"github.com/xrekson/game-backend/pkg/redis"
	"github.com/xrekson/game-backend/proto/world"
	"github.com/xrekson/game-backend/services/world/handler"
	"github.com/xrekson/game-backend/services/world/model"
	"github.com/xrekson/game-backend/services/world/presence"
	"github.com/xrekson/game-backend/services/world/repository"
	"github.com/xrekson/game-backend/services/world/service"
	"github.com/xrekson/game-backend/services/world/ws"
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

	redisClient, err := redis.NewRedisClient(cfg.Redis)
	if err != nil {
		logger.Error("Failed to connect to redis: %v", err)
	}

	producer := kafka.NewProducer(cfg.Kafka)
	defer producer.Close()

	mapRepo := repository.NewMapRepository(cassandraSession)
	presenceMgr := presence.NewPresenceManager(redisClient)
	worldSvc := service.NewWorldService(mapRepo, presenceMgr, producer)
	worldHdlr := handler.NewWorldHandler(worldSvc)

	// WebSocket Hub
	hub := ws.NewHub()
	go hub.Run()

	// WebSocket HTTP Server
	wsPort := cfg.Services.WorldWSPort
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		playerID := r.URL.Query().Get("player_id")
		if playerID == "" {
			playerID = "anon-player"
		}
		ws.ServeWS(hub, w, r, playerID, func(c *ws.Client, msg *model.WSMessage) {
			if msg.Type == "move" {
				payloadBytes, _ := json.Marshal(msg.Payload)
				var move model.MovePayload
				_ = json.Unmarshal(payloadBytes, &move)
				validated, isSafe, err := worldSvc.ValidateMovement(context.Background(), c.PlayerID, move.MapID, model.Position{X: move.X, Y: move.Y})
				if err == nil {
					respMsg, _ := json.Marshal(model.WSMessage{
						Type: "move_update",
						Payload: map[string]any{
							"player_id":    c.PlayerID,
							"position":     validated,
							"is_safe_zone": isSafe,
						},
					})
					c.Conn.WriteMessage(1, respMsg)
				}
			}
		})
	})

	go func() {
		logger.Info("World Service WebSocket server running on port %d", wsPort)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", wsPort), nil); err != nil {
			logger.Error("WS Server error: %v", err)
		}
	}()

	// gRPC Server
	grpcPort := cfg.Services.WorldGRPCPort
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", grpcPort))
	if err != nil {
		logger.Error("Failed to listen on gRPC port %d: %v", grpcPort, err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	worldpb.RegisterWorldServiceServer(grpcServer, worldHdlr)

	logger.Info("World Service gRPC server running on port %d", grpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		logger.Error("gRPC server failed: %v", err)
	}
}
