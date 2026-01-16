@echo off
echo 开始构建签到系统...

REM 检查Go环境
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo 错误: 未找到Go环境，请先安装Go
    pause
    exit /b 1
)

REM 清理旧的构建文件
echo 清理旧的构建文件...
if exist main.exe del main.exe
if exist main del main

REM 下载依赖
echo 下载依赖...
go mod download
go mod tidy

REM 构建应用
echo 构建应用...
go build -o main.exe main.go

if %errorlevel% equ 0 (
    echo 构建成功！
    echo 运行命令: main.exe
) else (
    echo 构建失败！
    pause
    exit /b 1
)

pause