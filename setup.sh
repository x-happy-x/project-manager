#!/usr/bin/env bash
set -euo pipefail

echo "==> Installing Project Manager (pm)"

REPO="x-happy-x/project-manager"
BINARY_NAME="pm"

OS_TYPE="linux"

ARCH="$(uname -m)"
case "$ARCH" in
    x86_64|amd64)   ARCH_TYPE="amd64" ;;
    arm64|aarch64)  ARCH_TYPE="arm64" ;;
    *)
        echo "==> Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

if ! command -v curl >/dev/null 2>&1; then
  echo "==> curl is required but not installed"
  exit 1
fi

if ! command -v mktemp >/dev/null 2>&1; then
  echo "==> mktemp is required but not installed"
  exit 1
fi

INSTALL_DIR="${PM_INSTALL_DIR:-$HOME/.local/bin}"

echo "==> Using install directory: $INSTALL_DIR"
mkdir -p "$INSTALL_DIR"

TEMP_DIR="$(mktemp -d)"
cleanup() {
  rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

cd "$TEMP_DIR"

echo "==> Detecting latest release for ${OS_TYPE}_${ARCH_TYPE}"
LATEST_URL="$(
  curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep -o "\"browser_download_url\": *\"[^\"]*pm_${OS_TYPE}_${ARCH_TYPE}[^\"]*\"" \
    | head -n1 \
    | cut -d '"' -f4
)"

if [ -z "$LATEST_URL" ]; then
  echo "==> Error: No prebuilt binary found for ${OS_TYPE}_${ARCH_TYPE}"
  exit 1
fi

echo "==> Downloading binary:"
echo "    $LATEST_URL"
curl -fsSL -o "$BINARY_NAME" "$LATEST_URL"
chmod +x "$BINARY_NAME"
BINARY_PATH="$TEMP_DIR/$BINARY_NAME"

echo "==> Installing pm to ${INSTALL_DIR}"
install -m 0755 "$BINARY_PATH" "${INSTALL_DIR}/pm"

# ---------------------------------------------------------------------
# PATH настройка: и в текущей сессии (если скрипт source'ят),
# и в профиль (bash/zsh)
# ---------------------------------------------------------------------

detect_profile_file() {
  if [ -n "${ZSH_VERSION-}" ]; then
    if [ -f "$HOME/.zshrc" ]; then
      echo "$HOME/.zshrc"
    elif [ -f "$HOME/.zprofile" ]; then
      echo "$HOME/.zprofile"
    else
      echo "$HOME/.zshrc"
    fi
  elif [ -n "${BASH_VERSION-}" ]; then
    if [ -f "$HOME/.bashrc" ]; then
      echo "$HOME/.bashrc"
    elif [ -f "$HOME/.bash_profile" ]; then
      echo "$HOME/.bash_profile"
    elif [ -f "$HOME/.profile" ]; then
      echo "$HOME/.profile"
    else
      echo "$HOME/.bashrc"
    fi
  else
    if [ -f "$HOME/.profile" ]; then
      echo "$HOME/.profile"
    else
      echo "$HOME/.profile"
    fi
  fi
}

PROFILE_FILE="$(detect_profile_file)"

# Добавляем INSTALL_DIR в PATH текущего процесса (если скрипт source'ят)
if ! printf '%s' ":$PATH:" | grep -q ":$INSTALL_DIR:"; then
  echo "==> Adding ${INSTALL_DIR} to PATH for current shell (if sourced)"
  export PATH="${INSTALL_DIR}:${PATH}"
fi

# Прописать в профиль, если ещё не прописан
if ! grep -q "$INSTALL_DIR" "$PROFILE_FILE" 2>/dev/null; then
  echo "==> Updating profile: $PROFILE_FILE"
  {
    echo ""
    echo "# Added by pm installer"
    echo "export PATH=\"$INSTALL_DIR:\$PATH\""
  } >> "$PROFILE_FILE"
else
  echo "==> PATH already configured in $PROFILE_FILE"
fi

echo "==> Running pm init"
"${INSTALL_DIR}/pm" init || true

echo ""
echo "==> Installation complete"
echo "    Binary: ${INSTALL_DIR}/pm"
echo ""
echo "    Quick start:"
echo "      pm add <path>     # Add a project"
echo "      pm list           # List all projects"
echo "      pm open <name>    # Open a project"
echo "      pm --help         # Show all commands"
echo ""
echo "==> If PATH does not update immediately, reload your shell:"
echo "    exec \$SHELL -l"