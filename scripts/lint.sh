#!/usr/bin/env bash
set -e

echo "==> Running linter..."

# Проверка наличия golangci-lint
if ! command -v golangci-lint &> /dev/null; then
    echo "golangci-lint not found, installing..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
fi

golangci-lint run --timeout=5m ./...

echo "==> Linting completed successfully!"
