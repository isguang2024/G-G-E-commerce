@echo off
setlocal
chcp 65001 >nul

echo ========================================
echo   Start frontend-fluentV2 (React + Vite)
echo ========================================

cd /d "%~dp0frontend-fluentV2"
if errorlevel 1 (
    echo Error: failed to enter "%~dp0frontend-fluentV2"
    pause
    exit /b 1
)

where node >nul 2>nul
if errorlevel 1 (
    echo Error: Node.js is required ^(>=20.19.0^)
    pause
    exit /b 1
)

node --version

where pnpm >nul 2>nul
if errorlevel 1 (
    echo Error: pnpm is required ^(>=8.8.0^)
    echo Install with: npm install -g pnpm
    pause
    exit /b 1
)

call pnpm --version

echo.
echo Starting frontend-fluentV2...
echo Working directory: %cd%
echo Default URL: http://localhost:5173
echo If the port is busy, Vite will pick another one.
echo.

call pnpm dev
set "EXIT_CODE=%errorlevel%"

if not "%EXIT_CODE%"=="0" (
    echo.
    echo frontend-fluentV2 exited with code: %EXIT_CODE%
    pause
)

exit /b %EXIT_CODE%
