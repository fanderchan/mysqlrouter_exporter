#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${script_dir}"

VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo dev)}"
COMMIT="${COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo unknown)}"
BUILD_DATE="${BUILD_DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"
GOOS="${GOOS:-$(go env GOOS)}"
GOARCH="${GOARCH:-$(go env GOARCH)}"
BIN_NAME="${BIN_NAME:-mysqlrouter_exporter}"

mkdir -p build dist

OUTPUT="build/${BIN_NAME}"
DIST_OUTPUT="dist/${BIN_NAME}-${GOOS}-${GOARCH}"

CGO_ENABLED=0 GOOS="${GOOS}" GOARCH="${GOARCH}" go build \
  -trimpath \
  -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildDate=${BUILD_DATE}" \
  -o "${OUTPUT}" .

cp "${OUTPUT}" "${DIST_OUTPUT}"
echo "built ${OUTPUT}"
echo "copied ${DIST_OUTPUT}"

