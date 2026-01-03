#!/bin/bash

REPO="huffmanks/config-stash"
VERSION=$(curl -s "https://api.github.com/repos/huffmanks/config-stash/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then
    ARCH="arm64"
fi

BINARY_NAME="config-stash-${OS}-${ARCH}"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}"

echo "Downloading ${BINARY_NAME}..."
curl -L -o config-stash "${DOWNLOAD_URL}"

chmod +x config-stash
sudo mv config-stash /usr/local/bin/config-stash

echo "Successfully installed! Type 'config-stash' to start."