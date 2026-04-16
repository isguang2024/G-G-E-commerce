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

echo Running backend migration...
echo.

go run ./cmd/migrate %*
set "RC=%errorlevel%"
echo.
echo Migration finished. code=%RC%
pause
exit /b %RC%
