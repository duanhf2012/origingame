@echo off
setlocal EnableExtensions

set "PROJECT_ROOT=%~dp0.."
cd /d "%PROJECT_ROOT%" || exit /b 1

where protoc >nul 2>nul || (
  echo protoc was not found in PATH. 1>&2
  exit /b 1
)
where protoc-gen-go >nul 2>nul || (
  echo protoc-gen-go was not found in PATH. 1>&2
  exit /b 1
)

for /f "delims=" %%V in ('protoc --version') do set "PROTOC_VERSION=%%V"
for /f "delims=" %%V in ('protoc-gen-go --version') do set "PROTOC_GEN_GO_VERSION=%%V"
if not "%PROTOC_VERSION%"=="libprotoc 24.0" (
  echo protoc 24.0 is required; found %PROTOC_VERSION%. 1>&2
  exit /b 1
)
if not "%PROTOC_GEN_GO_VERSION%"=="protoc-gen-go.exe v1.31.0" if not "%PROTOC_GEN_GO_VERSION%"=="protoc-gen-go v1.31.0" (
  echo protoc-gen-go 1.31.0 is required; found %PROTOC_GEN_GO_VERSION%. 1>&2
  exit /b 1
)

pushd "protocol\common" || exit /b 1
dir /b /a-d "*.proto" >nul 2>nul || (
  echo protocol/common contains no .proto files. 1>&2
  popd
  exit /b 1
)
for %%F in (*.proto) do (
  protoc --go_out=. --go_opt=paths=source_relative "%%F"
  if errorlevel 1 (
    popd
    exit /b 1
  )
)
popd

dir /b /a-d "protocol\rpc\*.proto" >nul 2>nul && (
  pushd "protocol\rpc" || exit /b 1
  for %%F in (*.proto) do (
    protoc --go_out=. --go_opt=paths=source_relative "%%F"
    if errorlevel 1 (
      popd
      exit /b 1
    )
  )
  popd
)

dir /b /a-d "protocol\rpc\*.go" >nul 2>nul && (
  go run github.com/duanhf2012/origin/v3/cmd/origingen rpc ./protocol/rpc
  if errorlevel 1 exit /b 1
)

go fmt ./protocol/...
exit /b %ERRORLEVEL%
