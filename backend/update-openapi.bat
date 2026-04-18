@echo off
setlocal EnableExtensions
set "RC=0"

chcp 65001 >nul
cd /d "%~dp0"

rem Regenerate all OpenAPI derived artifacts:
rem   bundle -> lint -> ogen (api/gen/) -> permission seed -> frontend error-codes.ts
rem
rem Usage:
rem   - After editing domains/*/paths.yaml or components/*.yaml: run this, restart backend
rem   - If DB schema/default data also changed: run cmd/migrate first, then this
rem   - Fresh database: cmd/migrate first, then this script

echo [api] bundling spec...
call npx --yes @redocly/cli bundle api/openapi/openapi.root.yaml -o api/openapi/dist/openapi.yaml --ext yaml
if errorlevel 1 goto :fail

echo [api] linting spec...
call npx --yes @redocly/cli lint api/openapi/dist/openapi.yaml --config api/openapi/redocly.yaml
if errorlevel 1 goto :fail

echo [api] running ogen...
go run github.com/ogen-go/ogen/cmd/ogen@latest --target api/gen --package gen --clean api/openapi/dist/openapi.yaml
if errorlevel 1 goto :fail

echo [api] generating permission seed + frontend error-codes.ts...
go run .\cmd\gen-permissions
if errorlevel 1 goto :fail

echo Done.
goto :pause

:fail
set "RC=%errorlevel%"
echo.
echo Failed. Check the output above.

:pause
echo Press any key to close this window.
pause >nul
endlocal & exit /b %RC%
