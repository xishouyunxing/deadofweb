#!/bin/bash

echo "🚀 签到系统环境配置脚本"
echo ======================

# 检查操作系统
OS="unknown"
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    OS="linux"
elif [[ "$OSTYPE" == "darwin"* ]]; then
    OS="macos"
elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]]; then
    OS="windows"
fi

echo "检测到操作系统: $OS"
echo ""

# 生成Session密钥
echo "🔐 生成Session密钥..."
if command -v openssl &> /dev/null; then
    SESSION_SECRET=$(openssl rand -base64 32 | tr '+/' '-_' | tr -d '=')
    echo "✅ 使用OpenSSL生成密钥成功"
elif [ -f /dev/urandom ]; then
    SESSION_SECRET=$(head -c 32 /dev/urandom | base64 | tr '+/' '-_' | tr -d '=')
    echo "✅ 使用 /dev/urandom 生成密钥成功"
else
    SESSION_SECRET=$(date +%s%N | sha256sum | head -c 32)
    echo "✅ 使用备用方法生成密钥成功"
fi

echo "密钥: $SESSION_SECRET"
echo ""

# 检查并创建.env文件
if [ -f ".env" ]; then
    echo "📝 .env文件已存在"
    echo "正在更新SESSION_SECRET..."
    
    # 更新SESSION_SECRET
    if grep -q "SESSION_SECRET=" .env; then
        sed -i "s/SESSION_SECRET=.*/SESSION_SECRET=$SESSION_SECRET/" .env
    else
        echo "" >> .env
        echo "# Session配置" >> .env
        echo "SESSION_SECRET=$SESSION_SECRET" >> .env
    fi
else
    echo "📝 创建新的.env文件..."
    cat > .env << EOF
# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=checkin_system

# Session配置
SESSION_SECRET=$SESSION_SECRET

# 邮件配置
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_EMAIL=your-email@gmail.com
SMTP_PASSWORD=your-app-password

# 服务器配置
SERVER_PORT=8080
EOF
fi

echo "✅ .env文件配置完成"
echo ""

# 下载Go依赖
echo "📦 下载Go依赖..."
if command -v go &> /dev/null; then
    go mod download
    go mod tidy
    echo "✅ 依赖下载完成"
else
    echo "❌ 未找到Go环境，请先安装Go 1.21+"
    exit 1
fi

echo ""
echo "🎉 环境配置完成！"
echo ""
echo "📋 接下来的步骤:"
echo "1. 编辑 .env 文件，配置数据库和邮件参数"
echo "2. 确保PostgreSQL服务正在运行"
echo "3. 运行应用: go run main.go"
echo "4. 访问: http://localhost:8080"
echo ""
echo "💡 重要提示:"
echo "- 请修改 DB_PASSWORD 为实际的数据库密码"
echo "- 请配置 SMTP_EMAIL 和 SMTP_PASSWORD"
echo "- 生产环境请使用更复杂的SESSION_SECRET"
echo ""