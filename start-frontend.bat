@echo off
chcp 65001 >nul

echo ========================================
echo   启动前端服务
echo ========================================

cd /d "%~dp0frontend"

where node >nul 2>nul
if %errorlevel% neq 0 (
    echo 错误: 未安装 Node.js，请先安装 Node.js ^(>=20.19.0^)
    pause
    exit /b 1
)

node --version

where pnpm >nul 2>nul
if %errorlevel% neq 0 (
    echo 错误: 未安装 pnpm，请先安装 pnpm ^(>=8.8.0^)
    echo 安装命令: npm install -g pnpm
    pause
    exit /b 1
)

pnpm --version

echo.
echo 正在启动前端服务...
echo 前端地址: http://localhost:5173
echo.

start cmd /k "cd /d C:\Users\Administrator\Documents\GitHub\G-G-E-commerce\frontend && pnpm dev"
