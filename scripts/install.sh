#!/usr/bin/env sh
set -eu

repo="${CONTROLI_REPO:-rgcsekaraa/controli}"
install_dir="${CONTROLI_INSTALL_DIR:-$HOME/.local/bin}"
os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m | tr '[:upper:]' '[:lower:]')"

case "$os" in
  darwin) os="darwin" ;;
  linux) os="linux" ;;
  *) echo "unsupported operating system: $os" >&2; exit 1 ;;
esac

case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  i386|i686) arch="386" ;;
  arm64|aarch64) arch="arm64" ;;
  armv6l) arch="armv6" ;;
  armv7l|armv8l) arch="armv7" ;;
  ppc64le) arch="ppc64le" ;;
  riscv64) arch="riscv64" ;;
  s390x) arch="s390x" ;;
  *) echo "unsupported architecture: $arch" >&2; exit 1 ;;
esac

asset="controli-$os-$arch"
url="https://github.com/$repo/releases/latest/download/$asset"
mkdir -p "$install_dir"
target="$install_dir/controli"

if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$url" -o "$target"
elif command -v wget >/dev/null 2>&1; then
  wget -q "$url" -O "$target"
else
  echo "curl or wget is required" >&2
  exit 1
fi

chmod +x "$target"
echo "installed $target"
echo "run: $target join <code>"
