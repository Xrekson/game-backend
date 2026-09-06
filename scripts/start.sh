#!/usr/bin/env bash
# Quick script to start the development server
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
"$SCRIPT_DIR/../dev.sh" start
