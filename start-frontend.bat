@echo off
setlocal
chcp 65001 >nul

echo ========================================
echo   启动前端服务
echo ========================================

cd /d "%~dp0frontend"
if errorlevel 1 (
    echo 错误: 无法进入前端目录 "%~dp0frontend"
    pause
    exit /b 1
)

where node >nul 2>nul
if errorlevel 1 (
    echo 错误: 未安装 Node.js，请先安装 Node.js ^(>=20.19.0^)
    pause
    exit /b 1
)

node --version

where pnpm >nul 2>nul
if errorlevel 1 (
    echo 错误: 未安装 pnpm，请先安装 pnpm ^(>=8.8.0^)
    echo 安装命令: npm install -g pnpm
    pause
    exit /b 1
)

call pnpm --version

echo.
echo 正在启动前端服务...
echo 前端目录: %cd%
echo 前端地址: http://localhost:5173
echo.

call pnpm dev
set "EXIT_CODE=%errorlevel%"

if not "%EXIT_CODE%"=="0" (
    echo.
    echo 前端启动失败，退出码: %EXIT_CODE%
    pause
)

exit /b %EXIT_CODE%
