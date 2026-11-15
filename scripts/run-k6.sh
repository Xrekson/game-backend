#!/usr/bin/env bash
# ==============================================================================
# Grafana k6 Load Test Runner for Distributed Game Backend
# Hardware Profile: 2 vCPU, 4GB RAM (via Linux cgroups v2)
# SLO: p95 < 0.75s (750ms), p99 < 2.0s (2000ms)
#
# Smart Scaling Tiers:
#   - VUs > 2,000       => 5 minutes (5m)
#   - 1,500 <= VUs <= 2,000 => 4 minutes (4m)
#   - 1,000 <= VUs < 1,500 => 3 minutes (3m)
#   - 500   <= VUs < 1,000 => 2 minutes (2m)
#   - 100   <= VUs < 500   => 1 minute  (1m)
#   - VUs < 100         => 30 seconds (30s)
#
# Usage:
#   ./scripts/run-k6.sh                # Interactive prompt with auto-scaled duration
#   ./scripts/run-k6.sh 2500           # 2,500 users => automatically runs for 5m
#   ./scripts/run-k6.sh 1800           # 1,800 users => automatically runs for 4m
#   ./scripts/run-k6.sh 50             # 50 users => runs for 30s
#   ./scripts/run-k6.sh 2000 1m        # Custom override: 2,000 users for 1m
# ==============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$BACKEND_DIR"
mkdir -p logs

# Raise file descriptor limit for high concurrency loads (avoids socket exhaustion)
ulimit -n 65535 2>/dev/null || ulimit -n 4096 2>/dev/null || true

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

function get_smart_duration() {
    local vus="$1"
    if [ "$vus" -gt 2000 ]; then
        echo "5m"
    elif [ "$vus" -ge 1500 ]; then
        echo "4m"
    elif [ "$vus" -ge 1000 ]; then
        echo "3m"
    elif [ "$vus" -ge 500 ]; then
        echo "2m"
    elif [ "$vus" -ge 100 ]; then
        echo "1m"
    else
        echo "30s"
    fi
}

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
    SMART_DUR=$(get_smart_duration "$VUS")
    DURATION="${2:-$SMART_DUR}"
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
        SMART_DUR=$(get_smart_duration "$VUS")
        read -p "  Enter test duration [auto-scaled default for $VUS VUs: $SMART_DUR]: " INPUT_DUR
        DURATION="${INPUT_DUR:-$SMART_DUR}"
    else
        VUS=50
        DURATION=$(get_smart_duration "$VUS")
    fi
fi

echo -e "\n\033[0;36m[2/2] Launching Grafana k6 Load Test...\033[0m"
echo -e "  • Concurrent Users (VUs): \033[1;32m$VUS\033[0m"
echo -e "  • Test Duration:          \033[1;32m$DURATION\033[0m"
echo -e "  • Target SLOs:            \033[0;34mp(95) < 750ms | p(99) < 2000ms\033[0m\n"

VUS="$VUS" DURATION="$DURATION" k6 run "$SCRIPT_DIR/k6/load_test.js"

if [ -f "logs/k6-report.md" ]; then
    echo -e "\n\033[0;32m================================================================================\033[0m"
    echo -e "\033[0;32m✓ k6 Benchmark Completed! Detailed Markdown Report saved to: logs/k6-report.md\033[0m"
    echo -e "\033[0;32m================================================================================\033[0m"
fi
