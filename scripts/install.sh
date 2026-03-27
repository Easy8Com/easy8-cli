#!/usr/bin/env bash
set -euo pipefail

REPO="Easy8Com/easy8-cli"
INSTALL_DIR="${EASY8_BIN_DIR:-$HOME/.local/bin}"
REQUESTED_VERSION="${EASY8_VERSION:-}"

if ! command -v curl >/dev/null 2>&1; then
  echo "curl is required but was not found." >&2
  exit 1
fi

OS_RAW="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH_RAW="$(uname -m)"

case "$OS_RAW" in
  linux) OS="linux" ;;
  darwin) OS="darwin" ;;
  *)
    echo "Unsupported OS: $OS_RAW" >&2
    exit 1
    ;;
esac

case "$ARCH_RAW" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH_RAW" >&2
    exit 1
    ;;
esac

if [ -n "$REQUESTED_VERSION" ]; then
  VERSION="$REQUESTED_VERSION"
else
  echo "Fetching latest version..."
  LATEST_URL="$(curl -fsSL -o /dev/null -w '%{url_effective}' "https://github.com/${REPO}/releases/latest")"
  VERSION="${LATEST_URL##*/}"
fi

if [ -z "$VERSION" ] || [ "$VERSION" = "latest" ] || [ "$VERSION" = "releases" ]; then
  echo "Failed to determine release version. Make sure a GitHub release exists." >&2
  exit 1
fi

BINARY_NAME="easy8-${OS}-${ARCH}"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}"
CHECKSUMS_URL="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

echo "Downloading ${BINARY_NAME} (${VERSION})..."
curl -fsSL "$DOWNLOAD_URL" -o "$TMP_DIR/$BINARY_NAME"
curl -fsSL "$CHECKSUMS_URL" -o "$TMP_DIR/checksums.txt"

EXPECTED="$(awk -v file="$BINARY_NAME" '$2 == file { print $1 }' "$TMP_DIR/checksums.txt")"
if [ -z "$EXPECTED" ]; then
  echo "Checksum entry for ${BINARY_NAME} not found in checksums.txt" >&2
  exit 1
fi

if command -v sha256sum >/dev/null 2>&1; then
  ACTUAL="$(sha256sum "$TMP_DIR/$BINARY_NAME" | awk '{print $1}')"
elif command -v shasum >/dev/null 2>&1; then
  ACTUAL="$(shasum -a 256 "$TMP_DIR/$BINARY_NAME" | awk '{print $1}')"
else
  echo "No SHA-256 tool found (need sha256sum or shasum)." >&2
  exit 1
fi

if [ "$EXPECTED" != "$ACTUAL" ]; then
  echo "Checksum mismatch for ${BINARY_NAME}" >&2
  echo "Expected: $EXPECTED" >&2
  echo "Actual:   $ACTUAL" >&2
  exit 1
fi

mkdir -p "$INSTALL_DIR"
TARGET_PATH="$INSTALL_DIR/easy8"
cp "$TMP_DIR/$BINARY_NAME" "$TARGET_PATH"
chmod +x "$TARGET_PATH"

echo
echo "easy8 ${VERSION} installed to ${TARGET_PATH}"

if ! printf '%s' "$PATH" | tr ':' '\n' | grep -Fxq "$INSTALL_DIR"; then
  echo
  echo "Add ${INSTALL_DIR} to your PATH:"
  SHELL_NAME="$(basename "${SHELL:-bash}")"
  case "$SHELL_NAME" in
    zsh)
      echo "  echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ~/.zshrc && source ~/.zshrc"
      ;;
    fish)
      echo "  fish_add_path ${INSTALL_DIR}"
      ;;
    *)
      echo "  echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ~/.bashrc && source ~/.bashrc"
      ;;
  esac
fi

echo
echo "Run 'easy8 setup' to configure Easy8 API access."
