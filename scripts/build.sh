#!/usr/bin/env bash
set -e

BINARY_NAME="project-manager"
VERSION="${VERSION:-dev}"

PLATFORMS=(
    "linux/amd64"
    "windows/amd64"
    "darwin/amd64"
)

echo "==> Building binaries for all platforms..."
echo "    Version: $VERSION"
echo ""

rm -rf dist
mkdir -p dist

for platform in "${PLATFORMS[@]}"; do
    GOOS="${platform%/*}"
    GOARCH="${platform#*/}"

    output_name="$BINARY_NAME"
    if [ "$GOOS" = "windows" ]; then
        output_name="${output_name}.exe"
    fi

    output_path="dist/${GOOS}_${GOARCH}/${output_name}"

    echo "==> Building for $GOOS/$GOARCH..."

    CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" \
        go build \
        -ldflags="-s -w -X main.version=$VERSION" \
        -o "$output_path" \
        ./cmd/pm-bin

    echo "    Created: $output_path"
done

echo ""
echo "==> Build completed successfully!"
echo ""
echo "Artifacts:"
find dist -type f -exec ls -lh {} \;
