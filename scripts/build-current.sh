#!/usr/bin/env bash
set -e

BINARY_NAME="project-manager"
VERSION="${VERSION:-dev}"

# Определение текущей платформы
GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)

output_name="$BINARY_NAME"
if [ "$GOOS" = "windows" ]; then
    output_name="${output_name}.exe"
fi

output_path="dist/${GOOS}_${GOARCH}/${output_name}"

echo "==> Building for current platform ($GOOS/$GOARCH)..."
echo "    Version: $VERSION"
echo ""

mkdir -p "dist/${GOOS}_${GOARCH}"

CGO_ENABLED=0 go build \
    -ldflags="-s -w -X main.version=$VERSION" \
    -o "$output_path" \
    ./cmd/pm-bin

echo "==> Build completed successfully!"
echo "    Binary: $output_path"
ls -lh "$output_path"
