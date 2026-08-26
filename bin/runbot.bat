@echo off
pushd "%~dp0.."

rem Start RobotService only. Start the target Login, Gateway, and Game services first.
go run ./cmd start --app-name origingame-robot --config ./config --pid-dir ./run --node test-robot-1

popd
