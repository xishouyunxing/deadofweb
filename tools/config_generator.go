package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"
)

func generateRandomString(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatal("生成随机数失败:", err)
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

func generateConfig() {
	jwtSecret := generateRandomString(32)
	
	// 创建.env配置文件内容
	configContent := fmt.Sprintf(`# 数据库配置
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=checkin_system

# JWT配置
JWT_SECRET=%s
JWT_EXPIRES_IN=24h

# 邮件配置
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_EMAIL=your-email@gmail.com
SMTP_PASSWORD=your-app-password

# 服务器配置
SERVER_PORT=8080`, jwtSecret)

	// 写入.env文件
	if err := os.WriteFile("../.env", []byte(configContent), 0644); err != nil {
		log.Fatal("写入.env文件失败:", err)
	}

	fmt.Println("✅ 配置文件生成成功！")
	fmt.Println()
	fmt.Println("JWT_SECRET已自动生成并设置:")
	fmt.Printf("%s\n", strings.Repeat("=", 50))
	fmt.Printf("%s\n", jwtSecret)
	fmt.Printf("%s\n", strings.Repeat("=", 50))
	fmt.Println()
	fmt.Println("📝 请注意修改以下配置:")
	fmt.Println("1. 数据库密码 (DB_PASSWORD)")
	fmt.Println("2. 邮件配置 (SMTP_EMAIL, SMTP_PASSWORD)")
	fmt.Println("3. 数据库名称 (如果需要) (DB_NAME)")
	fmt.Println()
	fmt.Println("⚠️  安全提醒:")
	fmt.Println("- 请将JWT_SECRET保存在安全的地方")
	fmt.Println("- 不要将.env文件提交到版本控制系统")
	fmt.Println("- 生产环境请使用强密码和加密连接")
}

func main() {
	fmt.Println("🔐 签到系统配置生成器")
	fmt.Println("====================")
	fmt.Println()

	// 检查是否已存在.env文件
	if _, err := os.Stat("../.env"); err == nil {
		fmt.Println("⚠️  .env文件已存在！")
		fmt.Print("是否要覆盖现有配置？(y/N): ")
		
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("操作已取消")
			return
		}
	}

	generateConfig()
}