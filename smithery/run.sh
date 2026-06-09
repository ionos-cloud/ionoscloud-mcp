#!/bin/sh
set -e

DIR="$(cd "$(dirname "$0")" && pwd)"

uname_s=$(uname -s 2>/dev/null || echo unknown)
case "$uname_s" in
  Darwin) os=darwin ;;
  Linux)  os=linux ;;
  *) echo "ionoscloud-mcp: unsupported OS: $uname_s" >&2; exit 1 ;;
esac

uname_m=$(uname -m 2>/dev/null || echo unknown)
case "$uname_m" in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) echo "ionoscloud-mcp: unsupported arch: $uname_m" >&2; exit 1 ;;
esac

bin="$DIR/bin/${os}_${arch}/ionoscloud-mcp"

if [ ! -f "$bin" ]; then
  echo "ionoscloud-mcp: binary not found for ${os}_${arch} at $bin" >&2
  exit 1
fi

chmod +x "$bin" 2>/dev/null || true

exec "$bin" "$@"
