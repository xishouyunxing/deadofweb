# 用户签到系统

一个基于Golang开发的用户签到系统网站，提供完整的用户注册、签到管理、提醒通知等功能。

## 功能特性

### 🎯 核心功能
- **用户注册**：支持用户名、邮箱注册和密码设置
- **用户登录**：基于JWT的安全认证系统
- **每日签到**：简单的签到功能，支持添加备注
- **签到历史**：查看详细的签到记录和统计数据

### 🔔 提醒功能
- **可配置提醒**：用户可开启/关闭签到提醒
- **灵活频率**：支持每日、按小时等多种提醒频率
- **智能提醒**：根据签到状态自动调整提醒时间
- **缺签处理**：连续两天未签到自动发送警告邮件

### 📧 邮件系统
- **模板化邮件**：支持自定义邮件模板
- **多种邮件类型**：欢迎邮件、提醒邮件、缺签警告等
- **实时发送**：基于事件触发的即时邮件通知

### ⏰ 定时任务
- **智能调度**：基于cron表达式的任务调度
- **自动检测**：定时检查签到状态和发送提醒
- **缺签监控**：每天早上8点自动检查缺签用户

## 技术架构

### 后端技术栈
- **Go 1.21+**：主要开发语言
- **Gin**：轻量级Web框架
- **GORM**：ORM数据库操作
- **PostgreSQL**：主数据库
- **Session**：用户认证（基于Cookie）
- **Cron**：定时任务调度
- **SMTP**：邮件发送

### 前端技术栈
- **Bootstrap 5**：响应式UI框架
- **JavaScript**：前端交互逻辑
- **HTML5/CSS3**：页面结构和样式

## 项目结构

```
checkin-system/
├── main.go                 # 主程序入口
├── go.mod                  # Go模块定义
├── .env                    # 环境变量配置
├── config/                 # 配置文件
│   ├── database.go        # 数据库配置
│   └── email.go           # 邮件配置
│   └── email_templates.json # 邮件模板
├── database/              # 数据库相关
│   └── database.go        # 数据库连接
├── models/                # 数据模型
│   ├── user.go           # 用户模型
│   ├── checkin.go        # 签到模型
│   └── reminder.go       # 提醒模型
├── handlers/             # 请求处理器
│   ├── user.go           # 用户相关
│   ├── checkin.go        # 签到相关
│   ├── reminder.go       # 提醒相关
│   └── pages.go          # 页面相关
├── services/             # 业务服务
│   ├── email.go          # 邮件服务
│   └── scheduler.go      # 定时任务服务
├── middleware/            # 中间件
│   └── auth.go           # 认证中间件
├── templates/            # HTML模板
│   ├── index.html        # 首页
│   ├── login.html        # 登录页
│   ├── register.html     # 注册页
│   └── dashboard.html    # 仪表板
├── static/               # 静态资源
│   └── css/
│       └── style.css     # 样式文件
└── README.md            # 项目文档
```

## 快速开始

### 1. 环境准备
- Go 1.21+
- PostgreSQL 12+
- SMTP邮件服务器

### 2. 配置环境变量

**方法一：自动配置（推荐）**
```bash
# Linux/macOS
chmod +x setup.sh
./setup.sh

# 或使用Go工具
cd tools && go run session_generator.go
```

**方法二：手动配置**
复制并编辑 `.env` 文件：
```bash
# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=checkin_system

# Session配置
SESSION_SECRET=your-session-secret-key-here

# 邮件配置
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_EMAIL=your-email@gmail.com
SMTP_PASSWORD=your-app-password

# 服务器配置
SERVER_PORT=8080
```

### 3. 安装依赖
```bash
go mod download
go mod tidy
```

### 4. 启动应用
```bash
go run main.go
```

### 5. 访问系统
打开浏览器访问：`http://localhost:8080`

## API接口

### 用户相关
- `POST /api/register` - 用户注册
- `POST /api/login` - 用户登录
- `POST /api/logout` - 用户登出
- `GET /api/profile` - 获取用户信息
- `PUT /api/profile` - 更新用户信息

### 签到相关
- `POST /api/checkin` - 用户签到
- `GET /api/checkin/history` - 获取签到历史
- `GET /api/checkin/status` - 获取签到状态

### 提醒相关
- `GET /api/reminder` - 获取提醒设置
- `PUT /api/reminder` - 更新提醒设置

## 数据库设计

### 用户表 (users)
- `id` - 主键
- `username` - 用户名（唯一）
- `email` - 邮箱（唯一）
- `password` - 密码（加密存储）
- `created_at` - 创建时间
- `updated_at` - 更新时间

### 签到表 (check_ins)
- `id` - 主键
- `user_id` - 用户ID（外键）
- `checkin_at` - 签到时间
- `note` - 签到备注
- `created_at` - 创建时间

### 提醒表 (check_in_reminders)
- `id` - 主键
- `user_id` - 用户ID（外键）
- `is_enabled` - 是否启用提醒
- `reminder_frequency` - 提醒频率（daily/hourly）
- `reminder_interval` - 提醒间隔（小时）
- `next_reminder` - 下次提醒时间
- `last_reminder` - 上次提醒时间

## 部署说明

### Docker部署
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/config ./config
EXPOSE 8080
CMD ["./main"]
```

### 系统服务部署
可以使用systemd或supervisor将应用部署为系统服务。

## 配置说明

### 邮件模板配置
邮件模板存储在 `config/email_templates.json` 文件中，支持以下类型：
- `welcome` - 欢迎邮件
- `daily_reminder` - 每日提醒
- `hourly_reminder` - 小时提醒
- `missed_checkin_warning` - 缺签警告

模板支持变量替换：
- `{{.Username}}` - 用户名
- `{{.Email}}` - 用户邮箱

## 安全特性

- **密码加密**：使用bcrypt加密用户密码
- **Session认证**：基于Cookie的安全会话管理
- **输入验证**：严格的输入参数验证
- **SQL注入防护**：GORM ORM防护
- **XSS防护**：前端输出转义
- **Cookie安全**：HttpOnly和Secure选项保护

## 监控和日志

系统内置了基本的日志记录功能：
- 应用启动日志
- 错误日志
- 定时任务执行日志
- 邮件发送状态日志

## 扩展开发

### 添加新的提醒类型
1. 在 `models/reminder.go` 中添加新的频率类型
2. 在 `services/scheduler.go` 中实现对应的提醒逻辑
3. 在前端添加相应的配置选项

### 自定义邮件模板
1. 编辑 `config/email_templates.json` 文件
2. 重启应用或调用重新加载接口

### 添加新的认证方式
1. 在 `middleware/auth.go` 中扩展认证逻辑
2. 在 `handlers/user.go` 中添加新的登录接口

## 常见问题

### Q: 如何修改邮件发送配置？
A: 编辑 `.env` 文件中的SMTP相关配置，确保邮件服务器地址、端口、账号密码正确。

### Q: 定时任务不执行怎么办？
A: 检查系统时间是否正确，确认cron表达式设置无误，查看应用日志排查错误。

### Q: 数据库连接失败？
A: 确认PostgreSQL服务正常运行，数据库连接参数正确，网络畅通。

## 许可证

本项目采用MIT许可证，详见LICENSE文件。

## 贡献指南

欢迎提交Issue和Pull Request来改进这个项目。在提交代码前，请确保：
1. 代码通过所有测试
2. 遵循项目的代码规范
3. 添加必要的注释和文档

## 联系方式

如有问题或建议，请通过以下方式联系：
- 提交Issue
- 发送邮件

---

感谢使用用户签到系统！