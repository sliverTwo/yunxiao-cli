#!/usr/bin/env sh
# Flow public-runner friendly CI bootstrap.
# Prefer Alibaba golang mirror (go.dev / proxy.golang.org SSL can be flaky on Flow runners).
# Parent may pipeline-update this script body into the live Flow YAML.
set -eu
cd "$(dirname "$0")/.."

GO_VER="${GO_VER:-1.24.4}"
GO_TGZ="go${GO_VER}.linux-amd64.tar.gz"
MIRROR_URL="https://mirrors.aliyun.com/golang/${GO_TGZ}"

if ! command -v go >/dev/null 2>&1 || [ "$(go env GOVERSION 2>/dev/null || true)" != "go${GO_VER}" ]; then
  echo "installing Go ${GO_VER} from Alibaba mirror..."
  curl -fsSL -o /tmp/go.tgz "$MIRROR_URL"
  rm -rf /usr/local/go
  tar -C /usr/local -xzf /tmp/go.tgz
  export PATH="/usr/local/go/bin:${PATH}"
fi

export GOPROXY="${GOPROXY:-https://goproxy.cn,direct}"
export GOSUMDB="${GOSUMDB:-sum.golang.google.cn}"

echo "go: $(go version)"
echo "GOPROXY=$GOPROXY"

VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo 0.15.2)}"
export VERSION
exec ./scripts/ci.sh
