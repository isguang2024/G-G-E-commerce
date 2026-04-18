@echo off
setlocal
chcp 65001 >nul

cd /d "%~dp0backend"
if errorlevel 1 (
  echo Failed to enter backend directory.
  pause
  exit /b 1
)

go version >nul 2>nul
if errorlevel 1 (
  echo Go is not installed or not in PATH.
  pause
  exit /b 1
)

if not exist "api\\gen\\oas_server_gen.go" (
  echo Missing generated backend OpenAPI files: backend\\api\\gen\\oas_server_gen.go
  echo Run backend\\update-openapi.bat first, then start the backend again.
  pause
  exit /b 1
)

echo Starting backend in debug mode...
echo URL: http://localhost:8080
echo.

go run cmd/server/main.go
set "RC=%errorlevel%"
echo.
echo Backend exited. code=%RC%
pause
exit /b %RC%
