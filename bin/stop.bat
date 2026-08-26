@echo off
pushd "%~dp0.."

go run ./cmd stop --app-name origingame-local --pid-dir ./run --timeout 2m10s

popd
