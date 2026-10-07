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
        echo "✨ stash ${VERSION} is already installed and up to date!"
        exit 0
    fi

    if [ "$FORCE_INSTALL" = true ]; then
        echo "⚡ Force install triggered. Reinstalling stash ${VERSION}..."
        echo
    else
        echo "🔄 Upgrading stash from ${CURRENT_VERSION} to ${VERSION}..."
        echo
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
    INSTALL_DIR="$HOME/.local/bin"
fi

FOUND_BINARIES=$(type -a -p stash 2>/dev/null || true)

if [ -n "$FOUND_BINARIES" ]; then
    echo "$FOUND_BINARIES" | while read -r legacy; do
        if [ -f "$legacy" ] && [ "$legacy" != "$INSTALL_DIR/stash" ]; then
            echo "🧹 Removing legacy binary from $legacy..."
            echo

            NEEDS_PASS=false
            if ! sudo -n true 2>/dev/null; then
                NEEDS_PASS=true
            fi

            sudo rm -f "$legacy" || true

            if [ "$NEEDS_PASS" = true ]; then
                echo
            fi
        fi
    done
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

if uname -a | grep -qE -i "android|debian|ubuntu" || { [ -f /etc/os-release ] && grep -qE -i "android|debian|ubuntu" /etc/os-release; }; then
    echo
    echo "📦 Debian/Ubuntu/Android environment detected. Configuring Zsh..."

    if ! command -v zsh >/dev/null 2>&1; then
        sudo DEBIAN_FRONTEND=noninteractive apt install -y zsh
    else
        echo
        echo "✨ Zsh is already installed. Skipping installation."
    fi

    if ! grep -q 'command -v zsh' ~/.bashrc 2>/dev/null; then
        cat << 'EOF' >> ~/.bashrc

if [ -x "$(command -v zsh)" ]; then
  export SHELL=$(command -v zsh)
  exec $(command -v zsh) -l
fi
EOF
    fi

    chsh -s $(which zsh) 2>/dev/null || true
fi

case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *)
    echo
    echo "⚠️  $INSTALL_DIR was not in your PATH. Adding it..."
    for rc in "$HOME/.zshrc" "$HOME/.bashrc"; do
        if [ -f "$rc" ] && ! grep -q "$INSTALL_DIR" "$rc"; then
            echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$rc"
        fi
    done
    export PATH="$INSTALL_DIR:$PATH"
    ;;
esac

echo
echo "✅ stash installed to $INSTALL_DIR/stash"

"$INSTALL_DIR/stash" --version

if [ -t 2 ] && command -v zsh >/dev/null 2>&1; then
    sleep 2
    echo
    echo "🔄 Reloading shell session..."
    echo
    exec zsh -l < /dev/tty
fi
