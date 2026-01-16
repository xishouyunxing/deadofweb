package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
)

func generateSecret(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		log.Fatal("生成随机数失败:", err)
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

func main() {
	fmt.Println("JWT密钥生成器")
	fmt.Println("================")
	fmt.Println()

	// 生成不同长度的密钥
	fmt.Printf("16字节密钥: %s\n", generateSecret(16))
	fmt.Printf("24字节密钥: %s\n", generateSecret(24))
	fmt.Printf("32字节密钥: %s\n", generateSecret(32))  // 推荐使用
	fmt.Printf("48字节密钥: %s\n", generateSecret(48))
	fmt.Printf("64字节密钥: %s\n", generateSecret(64))

	fmt.Println()
	fmt.Println("推荐使用32字节密钥，复制上面32字节密钥的值到.env文件中的JWT_SECRET字段")
}