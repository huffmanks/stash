#!/bin/bash

set -e

REPO="huffmanks/stash"

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
    CURRENT_VERSION=$(stash --version | awk '{print $NF}')
    if [ "$FORCE_INSTALL" = false ] && [ "${CURRENT_VERSION#v}" = "${VERSION#v}" ]; then
        echo "Stash ${VERSION} is already installed and up to date!"
        exit 0
    fi

    if [ "$FORCE_INSTALL" = true ]; then
        echo "Force install triggered. Reinstalling Stash ${VERSION}..."
    else
        echo "Upgrading Stash from ${CURRENT_VERSION} to ${VERSION}..."
    fi
fi

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
fi

if [ "$OS" = "darwin" ] && [ "$ARCH" = "amd64" ]; then
    PREFIX="stash-legacy"
else
    PREFIX="stash"
fi

BINARY_NAME="${PREFIX}_${VERSION#v}_${OS}_${ARCH}"

DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}.tar.gz"

echo "🚀 Downloading Stash ${VERSION} for ${OS}/${ARCH}..."
curl -L -o stash.tar.gz "${DOWNLOAD_URL}"

rm -f stash
tar -xzf stash.tar.gz stash
chmod +x stash

sudo mkdir -p /usr/local/bin
sudo mv -f stash /usr/local/bin/stash

if [ "$OS" = "darwin" ]; then
    sudo xattr -d com.apple.quarantine /usr/local/bin/stash 2>/dev/null || true
fi

rm stash.tar.gz

echo "✅ Stash installed to /usr/local/bin/stash"
stash --version
