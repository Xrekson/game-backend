#!/usr/bin/env bash
# ==============================================================================
# Grafana k6 Load Test Runner for Distributed Game Backend
# Hardware Profile: 2 vCPU, 4GB RAM (via Linux cgroups v2)
# SLO: p95 < 0.75s (750ms), p99 < 2.0s (2000ms)
#
# Usage:
#   ./scripts/run-k6.sh                # Prompts for amount of users & duration
#   ./scripts/run-k6.sh 50             # 50 users, default 30s
#   ./scripts/run-k6.sh 100 45s        # 100 users for 45 seconds
#   ./scripts/run-k6.sh quick          # Quick preset (20 users, 20s)
#   ./scripts/run-k6.sh smoke          # Smoke check (1 user, 5s)
# ==============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$BACKEND_DIR"
mkdir -p logs

# 1. Verify k6 is available
if ! command -v k6 >/dev/null 2>&1; then
    echo -e "\033[0;31m❌ Grafana k6 is not installed or not in PATH.\033[0m"
    echo -e "Please install k6 (e.g., 'sudo dnf install k6' or via your package manager)."
    exit 1
fi

echo -e "\033[0;36m[1/2] Checking API Gateway status (http://localhost:8080/health)...\033[0m"
if ! curl -s -f http://localhost:8080/health >/dev/null 2>&1; then
    echo -e "\033[1;33m  ! API Gateway is not responding at http://localhost:8080/health.\033[0m"
    echo -e "\033[1;33m  ! Please start your databases and backend first:\033[0m"
    echo -e "      podman-compose -f deployments/docker-compose.yml up -d"
    echo -e "      RATE_LIMIT_RPM=-1 ./dev.sh start-constrained"
    exit 1
fi
echo -e "\033[0;32m  ✓ API Gateway is online.\033[0m"

VUS=""
DURATION=""

# Determine user input
if [ "$1" = "smoke" ]; then
    VUS=1
    DURATION="5s"
elif [ "$1" = "quick" ]; then
    VUS=20
    DURATION="20s"
elif [[ "$1" =~ ^[0-9]+$ ]]; then
    VUS="$1"
    DURATION="${2:-30s}"
elif [ -n "$1" ]; then
    # Custom k6 flags passed directly
    k6 run "$SCRIPT_DIR/k6/load_test.js" "$@"
    exit 0
else
    # Interactive prompt for user input
    echo -e "\n\033[1;33m⚙️  Load Test Parameters:\033[0m"
    if [ -t 0 ]; then
        read -p "  Enter amount of concurrent users (VUs) [default: 50]: " INPUT_VUS
        VUS="${INPUT_VUS:-50}"
        read -p "  Enter test duration (e.g. 20s, 30s, 1m) [default: 30s]: " INPUT_DUR
        DURATION="${INPUT_DUR:-30s}"
    else
        VUS=50
        DURATION="30s"
    fi
fi

echo -e "\n\033[0;36m[2/2] Launching Grafana k6 Load Test...\033[0m"
echo -e "  • Concurrent Users (VUs): \033[1;32m$VUS\033[0m"
echo -e "  • Test Duration:          \033[1;32m$DURATION\033[0m"
echo -e "  • Target SLOs:            \033[0;34mp(95) < 750ms | p(99) < 2000ms\033[0m\n"

VUS="$VUS" DURATION="$DURATION" k6 run "$SCRIPT_DIR/k6/load_test.js"

if [ -f "logs/k6-report.md" ]; then
    echo -e "\n\033[0;32m================================================================================\033[0m"
    echo -e "\033[0;32m✓ k6 Benchmark Completed! Markdown Report saved to: logs/k6-report.md\033[0m"
    echo -e "\033[0;32m================================================================================\033[0m"
fi
