package cassandra

import (
	"fmt"
	"time"

	"github.com/gocql/gocql"
	"github.com/xrekson/game-backend/pkg/config"
	"github.com/xrekson/game-backend/pkg/logger"
)

type Session struct {
	*gocql.Session
}

func NewCassandraSession(cfg config.CassandraConfig) (*Session, error) {
	hosts := cfg.Hosts
	if len(hosts) == 0 {
		hosts = []string{"127.0.0.1"}
	}
	port := cfg.Port
	if port == 0 {
		port = 9042
	}
	keyspace := cfg.Keyspace
	if keyspace == "" {
		keyspace = "game_keyspace"
	}

	// 1. Initial connection without keyspace to create keyspace if not exists
	initCluster := gocql.NewCluster(hosts...)
	initCluster.Port = port
	initCluster.Timeout = 10 * time.Second
	initCluster.ConnectTimeout = 10 * time.Second
	if cfg.User != "" {
		initCluster.Authenticator = gocql.PasswordAuthenticator{
			Username: cfg.User,
			Password: cfg.Password,
		}
	}

	initSession, err := initCluster.CreateSession()
	if err != nil {
		logger.Warn("Initial Cassandra connection failed: %v", err)
	} else {
		createKeyspaceQuery := fmt.Sprintf(`
			CREATE KEYSPACE IF NOT EXISTS %s 
			WITH replication = {'class': 'SimpleStrategy', 'replication_factor': '1'};`, keyspace)
		if err := initSession.Query(createKeyspaceQuery).Exec(); err != nil {
			logger.Warn("Failed creating keyspace %s: %v", keyspace, err)
		}
		initSession.Close()
	}

	// 2. Connect cluster with configured keyspace
	cluster := gocql.NewCluster(hosts...)
	cluster.Port = port
	cluster.Keyspace = keyspace
	cluster.Consistency = gocql.One
	cluster.Timeout = 10 * time.Second
	cluster.ConnectTimeout = 10 * time.Second
	if cfg.User != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: cfg.User,
			Password: cfg.Password,
		}
	}

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("failed connecting to cassandra keyspace %s: %w", keyspace, err)
	}

	logger.Info("Connected to Cassandra successfully at %v:%d (keyspace: %s)", hosts, port, keyspace)

	return &Session{Session: session}, nil
}
