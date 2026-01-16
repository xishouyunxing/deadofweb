package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

func generateSessionSecret(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	// 使用URL安全的base64编码，并移除填充字符
	secret := base64.URLEncoding.EncodeToString(bytes)
	return strings.TrimRight(secret, "=")
}

func updateEnvFile(secret string) error {
	var envContent strings.Builder
	
	// 读取现有.env文件内容，如果存在的话
	if _, err := os.Stat("../.env"); err == nil {
		content, err := os.ReadFile("../.env")
		if err != nil {
			return fmt.Errorf("读取.env文件失败: %v", err)
		}
		lines := strings.Split(string(content), "\n")
		
		for _, line := range lines {
			if strings.HasPrefix(line, "SESSION_SECRET=") {
				continue // 跳过旧的SESSION_SECRET行
			}
			if line != "" {
				envContent.WriteString(line + "\n")
			}
		}
	} else {
		// 创建新的.env文件内容
		envContent.WriteString("# 数据库配置\n")
		envContent.WriteString("DB_HOST=localhost\n")
		envContent.WriteString("DB_PORT=5432\n")
		envContent.WriteString("DB_USER=postgres\n")
		envContent.WriteString("DB_PASSWORD=password\n")
		envContent.WriteString("DB_NAME=checkin_system\n")
		envContent.WriteString("\n")
		envContent.WriteString("# Session配置\n")
	}
	
	// 添加新的SESSION_SECRET
	envContent.WriteString(fmt.Sprintf("SESSION_SECRET=%s\n", secret))
	envContent.WriteString("\n")
	envContent.WriteString("# 邮件配置\n")
	envContent.WriteString("SMTP_HOST=smtp.gmail.com\n")
	envContent.WriteString("SMTP_PORT=587\n")
	envContent.WriteString("SMTP_EMAIL=your-email@gmail.com\n")
	envContent.WriteString("SMTP_PASSWORD=your-app-password\n")
	envContent.WriteString("\n")
	envContent.WriteString("# 服务器配置\n")
	envContent.WriteString("SERVER_PORT=8080\n")
	
	// 写入.env文件
	err := os.WriteFile("../.env", []byte(envContent.String()), 0644)
	if err != nil {
		return fmt.Errorf("写入.env文件失败: %v", err)
	}
	
	return nil
}

func main() {
	fmt.Println("🔐 Session密钥生成器 (Go版本)")
	fmt.Println("==============================")
	fmt.Println()

	// 生成不同长度的密钥
	secret16 := generateSessionSecret(16)
	secret24 := generateSessionSecret(24)
	secret32 := generateSessionSecret(32)
	secret48 := generateSessionSecret(48)

	fmt.Println("生成的Session密钥选项:")
	fmt.Println("16字节:", secret16)
	fmt.Println("24字节:", secret24)
	fmt.Println("32字节:", secret32, "⭐ 推荐")
	fmt.Println("48字节:", secret48)
	fmt.Println()

	// 选择推荐的32字节密钥
	selectedSecret := secret32

	fmt.Println("📝 正在更新.env文件...")
	if err := updateEnvFile(selectedSecret); err != nil {
		fmt.Printf("❌ 错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ .env文件更新成功！")
	fmt.Println()
	fmt.Println("📋 配置的SESSION_SECRET:")
	fmt.Println("==========================================")
	fmt.Printf("%s\n", selectedSecret)
	fmt.Println("==========================================")
	fmt.Println()

	fmt.Println("📝 接下来的步骤:")
	fmt.Println("1. 编辑 .env 文件，修改数据库和邮件配置")
	fmt.Println("2. 确保PostgreSQL服务正在运行")
	fmt.Println("3. 运行: go run main.go")
	fmt.Println("4. 访问: http://localhost:8080")
	fmt.Println()

	fmt.Println("💡 重要提示:")
	fmt.Println("- 请修改 DB_PASSWORD 为实际的数据库密码")
	fmt.Println("- 请配置 SMTP_EMAIL 和 SMTP_PASSWORD")
	fmt.Println("- 生产环境请考虑使用更强的安全配置")
}