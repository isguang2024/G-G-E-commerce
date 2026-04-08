@echo off
setlocal EnableExtensions

chcp 65001 >nul
cd /d "%~dp0"

rem This script only regenerates OpenAPI-derived seed data.
rem It does not replace database migrations.
rem
rem Typical use:
rem - After changing api/openapi/openapi.yaml: run this script, then restart backend.
rem - If schema/default-data changed too: run cmd/migrate as well.
rem - For a fresh database: run cmd/migrate first, then this script.

echo [1/2] Generate api/gen
go run github.com/ogen-go/ogen/cmd/ogen@latest --target api/gen --package gen --clean api/openapi/openapi.yaml
if errorlevel 1 goto :fail

echo [2/2] Generate permission seed
go run .\cmd\gen-permissions
if errorlevel 1 goto :fail

echo Done.
goto :pause

:fail
echo.
echo Failed. Check the output above.

:pause
echo Press any key to close this window.
pause >nul
endlocal
exit /b %errorlevel%
