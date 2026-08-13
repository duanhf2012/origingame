#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
PROJECT_ROOT=$(CDPATH= cd -- "$SCRIPT_DIR/.." && pwd)
NODE_ID=${1:-${ORIGINGAME_NODE:-login-1}}

cd "$PROJECT_ROOT"

exec go run ./cmd start \
  --app-name origingame \
  --config ./config \
  --pid-dir ./run \
  --node "$NODE_ID"
