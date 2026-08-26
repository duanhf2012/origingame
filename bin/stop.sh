#!/usr/bin/env sh
set -eu

cd "$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"

go run ./cmd stop --app-name origingame-local --pid-dir ./run --timeout 2m10s
