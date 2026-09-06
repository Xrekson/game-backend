#!/bin/bash
set -e

echo "Generating Protobuf files..."

if command -v protoc &> /dev/null; then
    protoc --go_out=. --go_opt=paths=source_relative \
        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
        proto/**/*.proto
    echo "Protobuf files compiled successfully with protoc."
else
    echo "protoc not found in PATH. Using pre-generated Go gRPC contracts."
fi
