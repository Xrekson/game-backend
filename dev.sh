#!/usr/bin/env bash

# ==============================================================================
# Development Server Management Script
# Usage:
#   ./dev.sh start   - Builds backend binaries, starts microservices & frontend
#   ./dev.sh stop    - Cleanly stops all microservices & frontend dev server
#   ./dev.sh status  - Displays running status and ports of all services
#   ./dev.sh restart - Restarts all microservices and frontend server
# ==============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$SCRIPT_DIR"
FRONTEND_DIR="$(cd "$SCRIPT_DIR/../template-angular-main" 2>/dev/null && pwd || echo "")"
PID_FILE="$BACKEND_DIR/.dev_pids"
LOG_DIR="$BACKEND_DIR/logs"

# Colors for UI output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

mkdir -p "$LOG_DIR" "$BACKEND_DIR/bin"

SERVICES=(
    "auth-service:50051:./cmd/auth-service"
    "player-service:50052:./cmd/player-service"
    "economy-service:50053:./cmd/economy-service"
    "inventory-service:50054:./cmd/inventory-service"
    "npc-service:50057:./cmd/npc-service"
    "trading-service:50055:./cmd/trading-service"
    "world-service:50056:./cmd/world-service"
    "analytics-consumer::./cmd/analytics-consumer"
    "gateway:8080:./cmd/gateway"
)

function check_docker() {
    echo -e "${CYAN}[1/4] Checking Database Infrastructure (Docker)...${NC}"
    if command -v docker >/dev/null 2>&1; then
        if docker compose -f "$BACKEND_DIR/deployments/docker-compose.yml" ps --services --filter "status=running" 2>/dev/null | grep -q "postgres"; then
            echo -e "${GREEN}  ✓ Docker containers (Postgres, Cassandra, Redis, Kafka) are running.${NC}"
        else
            echo -e "${YELLOW}  ! Docker infrastructure is not fully running. Attempting to start containers...${NC}"
            docker compose -f "$BACKEND_DIR/deployments/docker-compose.yml" up -d 2>/dev/null || echo -e "${YELLOW}  ! Could not auto-start docker containers. Please ensure PostgreSQL/Cassandra/Redis/Kafka are running.${NC}"
        fi
    elif command -v podman-compose >/dev/null 2>&1; then
        podman-compose -f "$BACKEND_DIR/deployments/docker-compose.yml" up -d 2>/dev/null || true
    else
        echo -e "${YELLOW}  ! 'docker' binary not found in PATH. Assuming databases are running natively or manually.${NC}"
    fi
}

function build_backend() {
    echo -e "${CYAN}[2/4] Building Backend Microservice Binaries...${NC}"
    cd "$BACKEND_DIR"
    for item in "${SERVICES[@]}"; do
        IFS=":" read -r name port path <<< "$item"
        echo -n "  Building $name... "
        go build -o "bin/$name" "$path"
        echo -e "${GREEN}DONE${NC}"
    done
}

function start_services() {
    if [ -f "$PID_FILE" ]; then
        echo -e "${YELLOW}Development server appears to be running already! Run './dev.sh stop' first.${NC}"
        status_services
        exit 1
    fi

    check_docker
    build_backend

    echo -e "${CYAN}[3/4] Launching Backend Microservices...${NC}"
    cd "$BACKEND_DIR"
    touch "$PID_FILE"

    for item in "${SERVICES[@]}"; do
        IFS=":" read -r name port path <<< "$item"
        logfile="$LOG_DIR/$name.log"
        ./bin/$name > "$logfile" 2>&1 &
        pid=$!
        echo "$pid:$name" >> "$PID_FILE"
        echo -e "  ${GREEN}✓${NC} Started ${PURPLE}$name${NC} (PID: ${YELLOW}$pid${NC}) -> Log: ${BLUE}logs/$name.log${NC}"
    done

    # Start Frontend Dev Server if available
    if [ -n "$FRONTEND_DIR" ] && [ -f "$FRONTEND_DIR/package.json" ]; then
        echo -e "${CYAN}[4/4] Starting Frontend Dev Server (Angular + Phaser)...${NC}"
        cd "$FRONTEND_DIR"
        logfile="$LOG_DIR/frontend.log"
        npm run dev > "$logfile" 2>&1 &
        fpid=$!
        echo "$fpid:frontend" >> "$PID_FILE"
        echo -e "  ${GREEN}✓${NC} Started ${PURPLE}frontend${NC} (PID: ${YELLOW}$fpid${NC}) -> http://localhost:4200"
    else
        echo -e "${YELLOW}[4/4] Frontend directory not found; skipping frontend launch.${NC}"
    fi

    echo -e "\n${GREEN}=======================================================${NC}"
    echo -e "${GREEN}🚀 Development server successfully started!${NC}"
    echo -e "  • ${CYAN}API Gateway${NC}:   http://localhost:8080"
    echo -e "  • ${CYAN}World WebSockets${NC}: ws://localhost:8081/ws"
    echo -e "  • ${CYAN}Frontend UI${NC}:     http://localhost:4200"
    echo -e "${GREEN}=======================================================${NC}"
}

function stop_services() {
    echo -e "${CYAN}Stopping Development Server...${NC}"
    if [ -f "$PID_FILE" ]; then
        while IFS=":" read -r pid name; do
            if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
                echo -n "  Stopping $name (PID $pid)... "
                kill "$pid" 2>/dev/null || true
                echo -e "${GREEN}STOPPED${NC}"
            fi
        done < "$PID_FILE"
        rm -f "$PID_FILE"
    fi

    # Cleanup any orphan processes matching service binary names or ng serve
    echo -e "  Cleaning up any remaining processes..."
    for item in "${SERVICES[@]}"; do
        IFS=":" read -r name port path <<< "$item"
        pkill -f "bin/$name" 2>/dev/null || true
    done
    pkill -f "ng serve" 2>/dev/null || true
    pkill -f "vite" 2>/dev/null || true

    echo -e "${GREEN}✓ All development servers cleanly stopped.${NC}"
}

function status_services() {
    echo -e "${CYAN}--- Development Server Status ---${NC}"
    if [ ! -f "$PID_FILE" ]; then
        echo -e "${YELLOW}Status: STOPPED (No active .dev_pids file found)${NC}"
        return
    fi

    printf "%-20s %-10s %-10s %-15s\n" "SERVICE" "PID" "STATUS" "LOG FILE"
    printf "%-20s %-10s %-10s %-15s\n" "-------" "---" "------" "--------"
    while IFS=":" read -r pid name; do
        if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
            status="${GREEN}RUNNING${NC}"
        else
            status="${RED}STOPPED${NC}"
        fi
        printf "%-20s %-10s %-20b %-15s\n" "$name" "$pid" "$status" "logs/$name.log"
    done < "$PID_FILE"
}

case "$1" in
    start)
        start_services
        ;;
    stop)
        stop_services
        ;;
    restart)
        stop_services
        sleep 1
        start_services
        ;;
    status)
        status_services
        ;;
    *)
        echo -e "Usage: $0 {start|stop|restart|status}"
        exit 1
        ;;
esac
