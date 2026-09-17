#!/usr/bin/env sh
# Minimal CI entrypoint for Codeup Flow / local hooks.
# Usage: ./scripts/ci.sh   OR   make ci
set -eu
cd "$(dirname "$0")/.."
VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo 0.15.2)}"
LDFLAGS="-X github.com/yunxiao-cli/yunxiao/internal/version.Version=${VERSION}"
go build -ldflags "$LDFLAGS" ./...
go test ./...
go vet ./...
