package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// EmailConfig 邮件配置
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPEmail    string
	SMTPPassword string
}

// GetEmailConfig 获取邮件配置
func GetEmailConfig() EmailConfig {
	return EmailConfig{
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnv("SMTP_PORT", "587"),
		SMTPEmail:    getEnv("SMTP_EMAIL", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
	}
}

// EmailTemplate 邮件模板
type EmailTemplate struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// EmailTemplates 邮件模板集合
type EmailTemplates map[string]EmailTemplate

// LoadEmailTemplates 加载邮件模板
func LoadEmailTemplates() (EmailTemplates, error) {
	configPath := filepath.Join("config", "email_templates.json")
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	
	var templates EmailTemplates
	err = json.Unmarshal(data, &templates)
	if err != nil {
		return nil, err
	}
	
	return templates, nil
}

// SaveEmailTemplates 保存邮件模板
func SaveEmailTemplates(templates EmailTemplates) error {
	configPath := filepath.Join("config", "email_templates.json")
	
	data, err := json.MarshalIndent(templates, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(configPath, data, 0644)
}