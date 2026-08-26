#!/usr/bin/env sh
set -eu

cd "$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"

go run ./cmd start --app-name origingame-local --config ./config --pid-dir ./run --node pub-discovery-1,pub-db-1,area1-db-1,area1-game-1,pub-gateway-1,pub-login-1
