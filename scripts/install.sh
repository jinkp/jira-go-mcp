#!/usr/bin/env sh
set -eu

OWNER="jinkp"
REPO="jira-go-mcp"
BIN_NAME="jira-mcp"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.mcp/bin}"
VERSION="${VERSION:-latest}"

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "error: required command not found: $1" >&2
    exit 1
  }
}

need_cmd curl
need_cmd tar
need_cmd mktemp
need_cmd uname

os=$(uname -s)
arch=$(uname -m)

case "$os" in
  Linux) os="linux" ;;
  Darwin) os="darwin" ;;
  *)
    echo "error: unsupported operating system: $os" >&2
    exit 1
    ;;
esac

case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  arm64|aarch64) arch="arm64" ;;
  *)
    echo "error: unsupported architecture: $arch" >&2
    exit 1
    ;;
esac

if [ "$VERSION" = "latest" ]; then
  VERSION=$(curl -fsSL "https://api.github.com/repos/$OWNER/$REPO/releases/latest" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)
  if [ -z "$VERSION" ]; then
    echo "error: could not resolve latest release version" >&2
    exit 1
  fi
fi

version_no_v=${VERSION#v}
asset="${REPO}_${version_no_v}_${os}_${arch}.tar.gz"
url="https://github.com/$OWNER/$REPO/releases/download/$VERSION/$asset"

tmpdir=$(mktemp -d)
trap 'rm -rf "$tmpdir"' EXIT INT TERM

archive="$tmpdir/$asset"

echo "Downloading $url"
curl -fL "$url" -o "$archive"

mkdir -p "$INSTALL_DIR"
tar -xzf "$archive" -C "$INSTALL_DIR"
chmod +x "$INSTALL_DIR/$BIN_NAME" || true

echo "Installed $BIN_NAME to $INSTALL_DIR/$BIN_NAME"
echo "If needed, add this to your shell profile:"
echo "export PATH=\"$INSTALL_DIR:\$PATH\""
