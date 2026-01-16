#!/bin/bash

# 构建脚本
echo "开始构建签到系统..."

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo "错误: 未找到Go环境，请先安装Go"
    exit 1
fi

# 清理旧的构建文件
echo "清理旧的构建文件..."
rm -f main

# 下载依赖
echo "下载依赖..."
go mod download
go mod tidy

# 构建应用
echo "构建应用..."
go build -o main main.go

if [ $? -eq 0 ]; then
    echo "构建成功！"
    echo "运行命令: ./main"
else
    echo "构建失败！"
    exit 1
fi