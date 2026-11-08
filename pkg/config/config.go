package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Env       string          `yaml:"env"`
	Services  ServicesConfig  `yaml:"services"`
	Postgres  PostgresConfig  `yaml:"postgres"`
	Cassandra CassandraConfig `yaml:"cassandra"`
	Redis     RedisConfig     `yaml:"redis"`
	Kafka     KafkaConfig     `yaml:"kafka"`
	JWT       JWTConfig       `yaml:"jwt"`
}

type ServicesConfig struct {
	GatewayPort       int `yaml:"gateway_port"`
	AuthGRPCPort      int `yaml:"auth_grpc_port"`
	PlayerGRPCPort    int `yaml:"player_grpc_port"`
	EconomyGRPCPort   int `yaml:"economy_grpc_port"`
	InventoryGRPCPort int `yaml:"inventory_grpc_port"`
	TradingGRPCPort   int `yaml:"trading_grpc_port"`
	WorldGRPCPort     int `yaml:"world_grpc_port"`
	WorldWSPort       int `yaml:"world_ws_port"`
	NPCGRPCPort       int `yaml:"npc_grpc_port"`
	RateLimitRPM      int `yaml:"rate_limit_rpm"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

func (p PostgresConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		p.User, p.Password, p.Host, p.Port, p.DBName, p.SSLMode)
}

type CassandraConfig struct {
	Hosts    []string `yaml:"hosts"`
	Port     int      `yaml:"port"`
	Keyspace string   `yaml:"keyspace"`
	User     string   `yaml:"user"`
	Password string   `yaml:"password"`
}

type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type KafkaConfig struct {
	Brokers []string `yaml:"brokers"`
	GroupID string   `yaml:"group_id"`
}

type JWTConfig struct {
	Secret            string `yaml:"secret"`
	AccessExpiryHours int    `yaml:"access_expiry_hours"`
}

func Load(path string) (*Config, error) {
	cfg := &Config{}
	
	if path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("failed to parse yaml config: %w", err)
			}
		}
	}

	overrideFromEnv(cfg)

	return cfg, nil
}

func overrideFromEnv(cfg *Config) {
	if env := os.Getenv("APP_ENV"); env != "" {
		cfg.Env = env
	}
	if host := os.Getenv("POSTGRES_HOST"); host != "" {
		cfg.Postgres.Host = host
	}
	if port := os.Getenv("POSTGRES_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Postgres.Port = p
		}
	}
	if user := os.Getenv("POSTGRES_USER"); user != "" {
		cfg.Postgres.User = user
	}
	if pass := os.Getenv("POSTGRES_PASSWORD"); pass != "" {
		cfg.Postgres.Password = pass
	}
	if db := os.Getenv("POSTGRES_DB"); db != "" {
		cfg.Postgres.DBName = db
	}
	if hosts := os.Getenv("CASSANDRA_HOSTS"); hosts != "" {
		cfg.Cassandra.Hosts = strings.Split(hosts, ",")
	}
	if port := os.Getenv("CASSANDRA_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Cassandra.Port = p
		}
	}
	if keyspace := os.Getenv("CASSANDRA_KEYSPACE"); keyspace != "" {
		cfg.Cassandra.Keyspace = keyspace
	}
	if rpm := os.Getenv("RATE_LIMIT_RPM"); rpm != "" {
		if r, err := strconv.Atoi(rpm); err == nil {
			cfg.Services.RateLimitRPM = r
		}
	}
	if host := os.Getenv("REDIS_HOST"); host != "" {
		cfg.Redis.Host = host
	}
	if port := os.Getenv("REDIS_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Redis.Port = p
		}
	}
	if pass, exists := os.LookupEnv("REDIS_PASSWORD"); exists {
		cfg.Redis.Password = pass
	}
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		cfg.Kafka.Brokers = strings.Split(brokers, ",")
	}
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		cfg.JWT.Secret = secret
	}
}
