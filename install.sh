#!/bin/bash

REPO="huffmanks/stash"
VERSION=$(curl -s "https://api.github.com/repos/huffmanks/stash/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
fi

BINARY_NAME="stash-${OS}-${ARCH}"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}"

echo "Downloading ${BINARY_NAME}..."
curl -L -o stash "${DOWNLOAD_URL}"

chmod +x stash
sudo mv stash /usr/local/bin/stash

echo "Successfully installed! Type 'stash' to start."