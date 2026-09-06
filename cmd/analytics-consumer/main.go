package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gocql/gocql"
	"github.com/xrekson/game-backend/pkg/cassandra"
	"github.com/xrekson/game-backend/pkg/config"
	"github.com/xrekson/game-backend/pkg/events"
	"github.com/xrekson/game-backend/pkg/kafka"
	"github.com/xrekson/game-backend/pkg/logger"
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
		logger.Warn("Cassandra connection notice in Analytics Consumer: %v", err)
	}

	if cassandraSession != nil {
		createTableCQL := `
			CREATE TABLE IF NOT EXISTS analytics_events (
				id timeuuid,
				topic text,
				event_key text,
				payload text,
				timestamp timestamp,
				created_at timestamp,
				PRIMARY KEY (topic, id)
			) WITH CLUSTERING ORDER BY (id DESC);`
		_ = cassandraSession.Query(createTableCQL).Exec()
	}

	topics := []string{
		events.TopicEconomyIncome,
		events.TopicEconomyUpgrades,
		events.TopicTradingCreated,
		events.TopicTradingCompleted,
		events.TopicWorldCombat,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	insertCQL := `INSERT INTO analytics_events (id, topic, event_key, payload, timestamp, created_at) VALUES (?, ?, ?, ?, ?, ?);`

	for _, topic := range topics {
		consumer := kafka.NewConsumer(cfg.Kafka, topic)
		go func(t string, c *kafka.Consumer) {
			defer c.Close()
			logger.Info("Started Analytics Consumer worker for topic: %s", t)
			for {
				select {
				case <-ctx.Done():
					return
				default:
					msg, err := c.ReadMessage(ctx)
					if err != nil {
						time.Sleep(100 * time.Millisecond)
						continue
					}

					if cassandraSession != nil {
						eventUUID := gocql.TimeUUID()
						err = cassandraSession.Query(insertCQL,
							eventUUID, t, string(msg.Key), string(msg.Value), msg.Time, time.Now(),
						).Exec()

						if err != nil {
							logger.Error("Failed inserting analytics record into Cassandra: %v", err)
						} else {
							logger.Info("Processed event from topic %s into Analytics Cassandra table", t)
						}
					}
				}
			}
		}(topic, consumer)
	}

	<-sigChan
	logger.Info("Shutting down Analytics Consumer Service gracefully...")
}
