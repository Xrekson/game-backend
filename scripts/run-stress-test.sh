#!/usr/bin/env bash
# ==============================================================================
# Stress Test Runner for 2-Core 4GB RAM Low-Spec Benchmark
# Usage:
#   ./scripts/run-stress-test.sh               # Runs full benchmark suite
#   ./scripts/run-stress-test.sh -c 50 -d 20s  # Custom concurrency & duration
#   ./scripts/run-stress-test.sh -mode health  # Run specific mode
# ==============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$BACKEND_DIR"

echo -e "\033[0;36m[1/3] Building stress-test tool...\033[0m"
mkdir -p bin logs
go build -o bin/stress-test ./cmd/stress-test
echo -e "\033[0;32m  ✓ Built bin/stress-test\033[0m"

echo -e "\033[0;36m[2/3] Checking API Gateway status (http://localhost:8080/health)...\033[0m"
if ! curl -s -f http://localhost:8080/health >/dev/null 2>&1; then
    echo -e "\033[1;33m  ! Gateway is not running on port 8080.\033[0m"
    echo -e "\033[1;33m  ! Please start your containers and backend first:\033[0m"
    echo -e "      podman-compose -f deployments/docker-compose.yml up -d"
    echo -e "      ./dev.sh start"
    exit 1
fi
echo -e "\033[0;32m  ✓ Gateway is online.\033[0m"

echo -e "\033[0;36m[3/3] Executing Stress Test Suite...\033[0m"
./bin/stress-test "$@"

echo -e "\n\033[0;32m================================================================================\033[0m"
echo -e "\033[0;32m✓ Benchmark complete! Report saved to logs/stress-test-report.md\033[0m"
echo -e "\033[0;32m================================================================================\033[0m"
