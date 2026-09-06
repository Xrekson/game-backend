# Distributed Incremental Dungeon Crawler & Trading Game - Backend

A high-performance, event-driven microservices backend for an incremental 2D top-down dungeon crawler game with income upgrades, equipment management, P2P marketplace trading, real-time WebSocket world interaction, and Kafka event streaming.

---

## 🏗️ Architecture Overview

```
                        ┌─────────────────┐
                        │   API Gateway   │  (REST + WebSockets, Auth, Rate-Limiting)
                        └────────┬────────┘
        ┌───────────┬────────────┼────────────┬───────────────┐
        ▼           ▼            ▼             ▼               ▼
   Auth Service  Player Svc   Economy Svc   Inventory Svc   World Svc (WS)
   (Postgres)    (Postgres)   (PG + Kafka)  (PG + Cassandra)(Redis+Cassandra)
                                    │             │               │
                                    └──────┬──────┴───────┬───────┘
                                           ▼               ▼
                                    Trading Service   NPC/SafeZone Svc
                                    (Postgres+Kafka)  (Cassandra)
                                           │
                                           ▼
                               ┌────────────────────┐
                               │   Kafka Cluster    │
                               │ (event backbone)   │
                               └──────────┬─────────┘
                                          ▼
                          Analytics/Audit Consumer → Cassandra (Time-Series)
```

### Microservices Breakdown

| Service | Port | Database / Message Bus | Purpose |
|---|---|---|---|
| **API Gateway** | `8080` (HTTP) | Redis (Rate limiting) | Unified REST router, JWT auth middleware, gRPC reverse proxy |
| **Auth Service** | `50051` (gRPC) | PostgreSQL (`users`) | Registration, login, password hashing (`bcrypt`), JWT tokens |
| **Player Service** | `50052` (gRPC) | PostgreSQL (`player_profiles`) | Profile stats, levels, XP progression |
| **Economy Service** | `50053` (gRPC) | PostgreSQL + Kafka | Passive income engine, utility upgrades, prestige resets, offline accumulation |
| **Inventory Service**| `50054` (gRPC) | PostgreSQL + Cassandra | Equipment ownership, item templates (Weapons, Armour, Armaments), slots |
| **Trading Service** | `50055` (gRPC) | PostgreSQL + Kafka | P2P marketplace listings, Transactional Outbox relay, Saga atomic trades |
| **World Service** | `50056` (gRPC), `8081` (WS) | Redis + Cassandra | WebSocket hub, 2D movement validation, dungeon combat resolution, presence |
| **NPC Service** | `50057` (gRPC) | Cassandra | Safe zone shop catalogs, NPC dialogues, shop purchases |
| **Analytics Consumer** | Background | Cassandra + Kafka | Event audit logging time-series store |

---

## 🚀 Prerequisites

Ensure the following tools are installed on your machine:
- **Go**: `v1.26+`
- **Docker & Docker Compose**: `v2.0+`
- **PostgreSQL**: `v16+` (or via Docker Compose)
- **Apache Cassandra**: `v4.1+` (or via Docker Compose)
- **Redis**: `v7.0+` (or via Docker Compose)
- **Apache Kafka + Zookeeper**: (or via Docker Compose)

---

## 🛠️ Basic Setup & Configuration

### 1. Configuration Files

Environment settings are defined in `configs/dev.yaml`:

```yaml
env: dev

services:
  gateway_port: 8080
  auth_grpc_port: 50051
  player_grpc_port: 50052
  economy_grpc_port: 50053
  inventory_grpc_port: 50054
  trading_grpc_port: 50055
  world_grpc_port: 50056
  world_ws_port: 8081
  npc_grpc_port: 50057

postgres:
  host: localhost
  port: 5432
  user: postgres
  password: postgrespassword
  dbname: game_db
  sslmode: disable

cassandra:
  hosts:
    - localhost
  port: 9042
  keyspace: game_keyspace

redis:
  host: localhost
  port: 6379

kafka:
  brokers:
    - localhost:9092
  group_id: game-backend-group
```

---

## 💻 How to Run Locally

### Step 1: Start Infrastructure Containers

Run Docker Compose to spin up local PostgreSQL, Cassandra, Redis, Zookeeper, and Kafka instances:

```bash
docker-compose -f deployments/docker-compose.yml up -d
```

### Step 2: Run Database Migrations & Cassandra Schemas

Apply PostgreSQL database tables and Cassandra CQL schemas:

```bash
# Apply PostgreSQL migrations
./scripts/migrate.sh

# Apply Cassandra CQL schema (or let services auto-create schema)
cqlsh -f scripts/cql_schema.cql
```

### Step 3: Build All Service Binaries

Compile all 9 microservice binaries into `bin/`:

```bash
mkdir -p bin
go build -o bin/auth-service ./cmd/auth-service
go build -o bin/player-service ./cmd/player-service
go build -o bin/economy-service ./cmd/economy-service
go build -o bin/inventory-service ./cmd/inventory-service
go build -o bin/npc-service ./cmd/npc-service
go build -o bin/trading-service ./cmd/trading-service
go build -o bin/world-service ./cmd/world-service
go build -o bin/gateway ./cmd/gateway
go build -o bin/analytics-consumer ./cmd/analytics-consumer
```

### Step 4: Run Microservices

Launch the microservices (in separate terminal sessions or background processes):

```bash
# 1. Start Core Services
./bin/auth-service &
./bin/player-service &
./bin/economy-service &
./bin/inventory-service &
./bin/npc-service &
./bin/trading-service &
./bin/world-service &

# 2. Start Analytics Consumer
./bin/analytics-consumer &

# 3. Start API Gateway
./bin/gateway
```

The API Gateway will be listening on **`http://localhost:8080`**, and the World WebSocket server on **`ws://localhost:8081/ws`**.

---

## 🧪 Running Unit Tests

Execute unit tests across all workspace packages:

```bash
go test ./pkg/auth/... ./services/economy/service/...
```
