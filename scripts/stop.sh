#!/usr/bin/env bash
# Quick script to stop the development server
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
"$SCRIPT_DIR/../dev.sh" stop
