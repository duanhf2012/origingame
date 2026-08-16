@echo off
setlocal

set "PROJECT_ROOT=%~dp0.."
set "NODE_ID=%~1"
if not defined NODE_ID set "NODE_ID=login-pub-1"

cd /d "%PROJECT_ROOT%" || exit /b 1

go run ./cmd start ^
  --app-name "origingame-%NODE_ID%" ^
  --config ./config ^
  --pid-dir ./run ^
  --node "%NODE_ID%"

exit /b %ERRORLEVEL%
