@echo off
chcp 65001 >nul

echo ========================================
echo   启动后端服务
echo ========================================

cd /d "%~dp0backend"

where go >nul 2>nul
if %errorlevel% neq 0 (
    echo 错误: 未安装 Go，请先安装 Go 语言环境
    pause
    exit /b 1
)

go version

echo.
echo 正在启动后端服务...
echo 后端地址: http://localhost:8080
echo.

start cmd /k "cd /d C:\Users\Administrator\Documents\GitHub\G-G-E-commerce\backend && go run cmd/server/main.go"
