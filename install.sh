#!/bin/bash

set -e
trap 'exit 1' INT TERM

REPO="huffmanks/stash"
APP_NAME="stash"
FORCE_INSTALL=false

for arg in "$@"; do
  case $arg in
    -f|--force)
      FORCE_INSTALL=true
      shift
      ;;
  esac
done

VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if command -v stash >/dev/null 2>&1; then
    CURRENT_VERSION=$(stash --version | grep -oE 'v?[0-9]+\.[0-9]+\.[0-9]+' | head -n1)
    if [ "$FORCE_INSTALL" = false ] && [ "${CURRENT_VERSION#v}" = "${VERSION#v}" ]; then
        echo "stash ${VERSION} is already installed and up to date!"
        exit 0
    fi

    if [ "$FORCE_INSTALL" = true ]; then
        echo "Force install triggered. Reinstalling stash ${VERSION}..."
    else
        echo "Upgrading stash from ${CURRENT_VERSION} to ${VERSION}..."
    fi
fi

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
fi

BINARY_NAME="${APP_NAME}_${VERSION#v}_${OS}_${ARCH}"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}.tar.gz"

if [ -z "$INSTALL_DIR" ]; then
    if [ "$(id -u)" -eq 0 ] || [ -w "/usr/local/bin" ]; then
        INSTALL_DIR="/usr/local/bin"
    else
        INSTALL_DIR="$HOME/.local/bin"
    fi
fi

mkdir -p "$INSTALL_DIR" 2>/dev/null || true

SUDO=""
if [ ! -w "$INSTALL_DIR" ]; then
    if command -v sudo >/dev/null 2>&1; then
        SUDO="sudo"
        $SUDO -v
    else
        echo "❌ Cannot write to $INSTALL_DIR and sudo is not available."
        exit 1
    fi
fi

TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"; exit 1' INT TERM
trap 'rm -rf "$TMP_DIR"' EXIT

echo "🚀 Downloading stash ${VERSION} for ${OS}/${ARCH}..."
curl -sSL -o "$TMP_DIR/stash.tar.gz" "$DOWNLOAD_URL"

tar -xzf "$TMP_DIR/stash.tar.gz" -C "$TMP_DIR" stash
chmod +x "$TMP_DIR/stash"

if ! $SUDO mv -f "$TMP_DIR/stash" "$INSTALL_DIR/stash"; then
    echo "❌ Installation failed."
    exit 1
fi

if [ "$OS" = "darwin" ]; then
    $SUDO xattr -d com.apple.quarantine "$INSTALL_DIR/stash" 2>/dev/null || true
fi

echo "✅ stash installed to $INSTALL_DIR/stash"

case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) echo "⚠️  $INSTALL_DIR is not in your PATH. Add it to your shell profile:"
     echo "   export PATH=\"\$PATH:$INSTALL_DIR\"" ;;
esac

"$INSTALL_DIR/stash" --version
