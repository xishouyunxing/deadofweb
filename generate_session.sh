#!/bin/bash

echo "🔐 Session密钥生成器"
echo ===================

# 方法1: 使用openssl生成随机字符串
if command -v openssl &> /dev/null; then
    echo "使用OpenSSL生成Session密钥..."
    SESSION_SECRET=$(openssl rand -base64 32 | tr '+/' '-_' | tr -d '=')
# 方法2: 使用 /dev/urandom
elif [ -f /dev/urandom ]; then
    echo "使用 /dev/urandom 生成Session密钥..."
    SESSION_SECRET=$(head -c 32 /dev/urandom | base64 | tr '+/' '-_' | tr -d '=')
# 方法3: 使用date和随机数
else
    echo "使用备用方法生成Session密钥..."
    SESSION_SECRET=$(date +%s%N | sha256sum | head -c 32)
fi

echo ""
echo "✅ Session密钥生成成功！"
echo ""
echo "生成的密钥 (32字节):"
echo "========================================"
echo "$SESSION_SECRET"
echo "========================================"
echo ""
echo "请将此密钥复制到.env文件的SESSION_SECRET字段中"
echo ""

# 检查.env文件是否存在
if [ -f ".env" ]; then
    echo "当前.env文件中的SESSION_SECRET设置:"
    grep "SESSION_SECRET" .env || echo "⚠️  SESSION_SECRET未设置"
else
    echo "⚠️  .env文件不存在，将创建新的.env文件"
    echo ""
    echo "正在创建.env文件..."
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
    echo "✅ .env文件创建完成！"
fi

echo ""
echo "📝 接下来的步骤:"
echo "1. 修改.env文件中的数据库和邮件配置"
echo "2. 确保SESSION_SECRET字段已设置为上面的值"
echo "3. 运行: go run main.go"
echo ""
echo "💡 提示: 如果您想重新生成密钥，可以再次运行此脚本"
echo ""