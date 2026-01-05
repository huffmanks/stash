#!/bin/bash

set -e

REPO="huffmanks/stash"
VERSION=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if command -v stash >/dev/null 2>&1; then
    CURRENT_VERSION=$(stash --version | awk '{print $NF}')
    if [ "${CURRENT_VERSION#v}" = "${VERSION#v}" ]; then
        echo "✨ Stash ${VERSION} is already installed and up to date!"
        exit 0
    fi
    echo "upgrading Stash from ${CURRENT_VERSION} to ${VERSION}..."
fi

OS=$(uname -s | tr '[:upper:]' '[:lower:]')

ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
fi

BINARY_NAME="stash_${VERSION#v}_${OS}_${ARCH}"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}.tar.gz"

echo "🚀 Downloading Stash ${VERSION} for ${OS}/${ARCH}..."
curl -L -o stash.tar.gz "${DOWNLOAD_URL}"

tar -xzf stash.tar.gz stash
chmod +x stash
sudo mv stash /usr/local/bin/stash
rm stash.tar.gz

echo "✅ Successfully installed! Type 'stash' to start."
