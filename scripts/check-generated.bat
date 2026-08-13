@echo off
setlocal EnableExtensions EnableDelayedExpansion

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

set "TEMP_ROOT=%TEMP%\origingame-proto-check-%RANDOM%-%RANDOM%"
if exist "%TEMP_ROOT%\" (
  echo Temporary directory already exists: %TEMP_ROOT%. 1>&2
  exit /b 1
)
mkdir "%TEMP_ROOT%" || exit /b 1
mkdir "%TEMP_ROOT%\common" || goto :failure
mkdir "%TEMP_ROOT%\rpc" || goto :failure

pushd "protocol\common" || goto :failure
dir /b /a-d "*.proto" >nul 2>nul || (
  echo protocol/common contains no .proto files. 1>&2
  popd
  goto :failure
)
for %%F in (*.proto) do (
  protoc --go_out="%TEMP_ROOT%\common" --go_opt=paths=source_relative "%%F"
  if errorlevel 1 (
    popd
    goto :failure
  )
)

for %%F in (*.proto) do (
  if not exist "%%~nF.pb.go" (
    echo protocol/common/%%~nF.pb.go is missing; run scripts/generate.bat. 1>&2
    popd
    goto :failure
  )
  fc /b "%TEMP_ROOT%\common\%%~nF.pb.go" "%%~nF.pb.go" >nul
  if errorlevel 1 (
    echo protocol/common/%%~nF.pb.go is stale; run scripts/generate.bat. 1>&2
    popd
    goto :failure
  )
)
for %%F in (*.pb.go) do (
  set "GENERATED_BASE=%%~nF"
  set "PROTO_NAME=!GENERATED_BASE:.pb=!.proto"
  if not exist "!PROTO_NAME!" (
    echo protocol/common/%%F has no matching .proto source. 1>&2
    popd
    goto :failure
  )
)
popd

dir /b /a-d "protocol\rpc\*.proto" >nul 2>nul && (
  pushd "protocol\rpc" || goto :failure
  for %%F in (*.proto) do (
    protoc --go_out="%TEMP_ROOT%\rpc" --go_opt=paths=source_relative "%%F"
    if errorlevel 1 (
      popd
      goto :failure
    )
  )
  for %%F in (*.proto) do (
    if not exist "%%~nF.pb.go" (
      echo protocol/rpc/%%~nF.pb.go is missing; run scripts/generate.bat. 1>&2
      popd
      goto :failure
    )
    fc /b "%TEMP_ROOT%\rpc\%%~nF.pb.go" "%%~nF.pb.go" >nul
    if errorlevel 1 (
      echo protocol/rpc/%%~nF.pb.go is stale; run scripts/generate.bat. 1>&2
      popd
      goto :failure
    )
  )
  for %%F in (*.pb.go) do (
    set "GENERATED_BASE=%%~nF"
    set "PROTO_NAME=!GENERATED_BASE:.pb=!.proto"
    if not exist "!PROTO_NAME!" (
      echo protocol/rpc/%%F has no matching .proto source. 1>&2
      popd
      goto :failure
    )
  )
  popd
)

dir /b /a-d "protocol\rpc\*.go" >nul 2>nul && (
  go run github.com/duanhf2012/origin/v3/cmd/origingen rpc --check ./protocol/rpc
  if errorlevel 1 goto :failure
)

rmdir /s /q "%TEMP_ROOT%"
exit /b 0

:failure
if exist "%TEMP_ROOT%\" rmdir /s /q "%TEMP_ROOT%"
exit /b 1
