#!/bin/bash
set -e

echo "🚀 Installing Project Manager (pm)..."

# Detect OS
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
    Linux*)     OS_TYPE="linux";;
    Darwin*)    OS_TYPE="darwin";;
    CYGWIN*|MINGW*|MSYS*) OS_TYPE="windows";;
    *)          echo "❌ Unsupported OS: $OS"; exit 1;;
esac

case "$ARCH" in
    x86_64|amd64)   ARCH_TYPE="amd64";;
    arm64|aarch64)  ARCH_TYPE="arm64";;
    *)              echo "❌ Unsupported architecture: $ARCH"; exit 1;;
esac

# Download the latest release
REPO="yourusername/project-manager"
BINARY_NAME="pm"
if [ "$OS_TYPE" = "windows" ]; then
    BINARY_NAME="pm.exe"
fi

echo "📦 Detecting latest release..."
LATEST_URL=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep "browser_download_url.*${OS_TYPE}_${ARCH_TYPE}" | cut -d '"' -f 4)

if [ -z "$LATEST_URL" ]; then
    echo "⚠️  No pre-built binary found. Building from source..."

    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        echo "❌ Go is not installed. Please install Go 1.21+ from https://golang.org/dl/"
        exit 1
    fi

    # Clone and build
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    git clone "https://github.com/$REPO.git"
    cd project-manager
    go build -o "$BINARY_NAME" ./cmd/pm
    BINARY_PATH="$TEMP_DIR/project-manager/$BINARY_NAME"
else
    echo "⬇️  Downloading from $LATEST_URL..."
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    curl -L -o "$BINARY_NAME" "$LATEST_URL"
    chmod +x "$BINARY_NAME"
    BINARY_PATH="$TEMP_DIR/$BINARY_NAME"
fi

# Install binary
if [ "$OS_TYPE" = "darwin" ] || [ "$OS_TYPE" = "linux" ]; then
    INSTALL_DIR="/usr/local/bin"

    if [ -w "$INSTALL_DIR" ]; then
        mv "$BINARY_PATH" "$INSTALL_DIR/$BINARY_NAME"
    else
        echo "🔐 Need sudo access to install to $INSTALL_DIR"
        sudo mv "$BINARY_PATH" "$INSTALL_DIR/$BINARY_NAME"
    fi

    echo "✅ Installed to $INSTALL_DIR/$BINARY_NAME"
elif [ "$OS_TYPE" = "windows" ]; then
    INSTALL_DIR="$HOME/bin"
    mkdir -p "$INSTALL_DIR"
    mv "$BINARY_PATH" "$INSTALL_DIR/$BINARY_NAME"

    # Add to PATH if not already there
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        echo "📝 Add this line to your shell profile (~/.bashrc or ~/.zshrc):"
        echo "   export PATH=\"\$PATH:$INSTALL_DIR\""
    fi

    echo "✅ Installed to $INSTALL_DIR/$BINARY_NAME"
fi

# Initialize pm
echo "⚙️  Initializing pm..."
pm init

echo ""
echo "🎉 Installation complete!"
echo ""
echo "Quick start:"
echo "  pm add <path>          # Add a project"
echo "  pm list                # List all projects"
echo "  pm open <name>         # Open a project"
echo "  pm --help              # Show all commands"
