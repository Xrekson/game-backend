#!/usr/bin/env bash
# ==============================================================================
# Grafana k6 Load Test Runner for Distributed Game Backend
# Hardware Profile: 2 vCPU, 4GB RAM (via Linux cgroups v2)
# SLO: p95 < 0.75s (750ms), p99 < 2.0s (2000ms)
#
# Usage:
#   ./scripts/run-k6.sh            # Standard ramping load test (up to 75 VUs, ~1m)
#   ./scripts/run-k6.sh quick      # Quick 20s test (20 VUs)
#   ./scripts/run-k6.sh smoke      # 5s smoke test (1 VU)
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

echo -e "\033[0;36m[2/2] Launching Grafana k6 Load Test...\033[0m"
echo -e "\033[0;34m  • SLO Thresholds: p(95) < 750ms (0.75s) | p(99) < 2000ms (2.0s)\033[0m"

MODE="$1"

if [ "$MODE" = "smoke" ]; then
    echo -e "\033[0;35m  • Running Smoke Test (1 VU, 5s)...\033[0m"
    k6 run --vus 1 --duration 5s "$SCRIPT_DIR/k6/load_test.js"
elif [ "$MODE" = "quick" ]; then
    echo -e "\033[0;35m  • Running Quick Benchmark (20 VUs, 20s)...\033[0m"
    k6 run --vus 20 --duration 20s "$SCRIPT_DIR/k6/load_test.js"
else
    # Standard full load test with ramping stages
    k6 run "$SCRIPT_DIR/k6/load_test.js" "$@"
fi

if [ -f "logs/k6-report.md" ]; then
    echo -e "\n\033[0;32m================================================================================\033[0m"
    echo -e "\033[0;32m✓ k6 Benchmark Completed! Markdown Report saved to: logs/k6-report.md\033[0m"
    echo -e "\033[0;32m================================================================================\033[0m"
fi
