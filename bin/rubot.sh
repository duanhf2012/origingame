#!/usr/bin/env sh
set -eu

cd "$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"

# 只启动 RobotService；请先独立启动被测 Login、Gateway、Game 等服务。
go run ./cmd start --app-name origingame-robot --config ./config --pid-dir ./run --node test-robot-1
