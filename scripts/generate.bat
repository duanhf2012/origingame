@echo off
setlocal EnableExtensions

set "PROJECT_ROOT=%~dp0.."
cd /d "%PROJECT_ROOT%" || exit /b 1

rem Resolve protoc and protoc-gen-go. Prefer locally provisioned tools under scripts\tools;
rem fall back to PATH for environments that already have the exact versions.
rem PROTOC/PROTOC_GEN_GO values are stored WITH inner double quotes
rem (e.g. set "PROTOC="C:\path\protoc.exe""), so paths containing spaces or
rem non-ASCII characters expand correctly both in `for /f` and in direct calls.
set "TOOLS_DIR=%~dp0tools"
set "PROTOC="
set "PROTOC_GEN_GO="
if exist "%TOOLS_DIR%\protoc.exe" set "PROTOC="%TOOLS_DIR%\protoc.exe""
if exist "%TOOLS_DIR%\protoc-gen-go.exe" set "PROTOC_GEN_GO="%TOOLS_DIR%\protoc-gen-go.exe""
if not defined PROTOC (
  where protoc >nul 2>nul && set "PROTOC=protoc"
)
if not defined PROTOC_GEN_GO (
  where protoc-gen-go >nul 2>nul && set "PROTOC_GEN_GO=protoc-gen-go"
)
if not defined PROTOC (
  echo protoc was not found in PATH, and %TOOLS_DIR%\protoc.exe does not exist. 1>&2
  echo Put protoc 24.0 win64 into %TOOLS_DIR%, or add protoc to PATH and retry. 1>&2
  exit /b 1
)
if not defined PROTOC_GEN_GO (
  echo protoc-gen-go was not found in PATH, and %TOOLS_DIR%\protoc-gen-go.exe does not exist. 1>&2
  echo Put protoc-gen-go v1.31.0 win64 into %TOOLS_DIR%, or add it to PATH and retry. 1>&2
  exit /b 1
)

for /f "delims=" %%V in ('%PROTOC% --version') do set "PROTOC_VERSION=%%V"
for /f "delims=" %%V in ('%PROTOC_GEN_GO% --version') do set "PROTOC_GEN_GO_VERSION=%%V"
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
  %PROTOC% --go_out=. --go_opt=paths=source_relative "%%F"
  if errorlevel 1 (
    popd
    exit /b 1
  )
)
popd

dir /b /a-d "protocol\rpc\*.proto" >nul 2>nul && (
  pushd "protocol\rpc" || exit /b 1
  for %%F in (*.proto) do (
    %PROTOC% --go_out=. --go_opt=paths=source_relative "%%F"
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
