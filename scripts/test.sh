#!/usr/bin/env bash
set -e

echo "==> Running tests..."

# Запуск тестов с race detector и coverage
go test -v -race -coverprofile=coverage.out ./...

echo ""
echo "==> Test coverage:"
go tool cover -func=coverage.out | tail -n 1

echo ""
echo "==> Tests completed successfully!"
echo "Coverage report saved to: coverage.out"
echo "To view HTML coverage report, run: go tool cover -html=coverage.out"
