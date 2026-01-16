.PHONY: help setup build run clean test deps

# 默认目标
help:
	@echo "签到系统管理命令："
	@echo ""
	@echo "  setup     - 配置环境（生成密钥和.env文件）"
	@echo "  deps      - 下载依赖"
	@echo "  build     - 构建应用"
	@echo "  run       - 运行应用"
	@echo "  clean     - 清理构建文件"
	@echo "  test      - 运行测试"
	@echo ""
	@echo "快速开始:"
	@echo "  make setup && make run"

# 配置环境
setup:
	@echo "🔐 生成Session密钥..."
	@if command -v openssl >/dev/null 2>&1; then \
		SECRET=$$(openssl rand -base64 32 | tr '+/' '-_' | tr -d '='); \
	elif [ -f /dev/urandom ]; then \
		SECRET=$$(head -c 32 /dev/urandom | base64 | tr '+/' '-_' | tr -d '='); \
	else \
		SECRET=$$(date +%s%N | sha256sum | head -c 32); \
	fi; \
	echo "✅ 密钥生成成功: $$SECRET"; \
	if [ -f .env ]; then \
		sed -i.bak '/^SESSION_SECRET=/d' .env; \
		echo "SESSION_SECRET=$$SECRET" >> .env; \
		rm -f .env.bak; \
	else \
		echo "# 数据库配置" > .env; \
		echo "DB_HOST=localhost" >> .env; \
		echo "DB_PORT=5432" >> .env; \
		echo "DB_USER=postgres" >> .env; \
		echo "DB_PASSWORD=password" >> .env; \
		echo "DB_NAME=checkin_system" >> .env; \
		echo "" >> .env; \
		echo "# Session配置" >> .env; \
		echo "SESSION_SECRET=$$SECRET" >> .env; \
		echo "" >> .env; \
		echo "# 邮件配置" >> .env; \
		echo "SMTP_HOST=smtp.gmail.com" >> .env; \
		echo "SMTP_PORT=587" >> .env; \
		echo "SMTP_EMAIL=your-email@gmail.com" >> .env; \
		echo "SMTP_PASSWORD=your-app-password" >> .env; \
		echo "" >> .env; \
		echo "# 服务器配置" >> .env; \
		echo "SERVER_PORT=8080" >> .env; \
	fi; \
	echo "✅ .env文件配置完成"

# 下载依赖
deps:
	@echo "📦 下载Go依赖..."
	go mod download
	go mod tidy

# 构建应用
build:
	@echo "🔨 构建应用..."
	go build -o main main.go

# 运行应用
run: deps
	@echo "🚀 启动应用..."
	go run main.go

# 清理构建文件
clean:
	@echo "🧹 清理构建文件..."
	rm -f main main.exe

# 运行测试
test:
	@echo "🧪 运行测试..."
	go test ./...

# Docker构建
docker-build:
	@echo "🐳 构建Docker镜像..."
	docker build -t checkin-system .

# Docker运行
docker-run:
	@echo "🐳 运行Docker容器..."
	docker-compose up -d

# Docker停止
docker-stop:
	@echo "🐳 停止Docker容器..."
	docker-compose down