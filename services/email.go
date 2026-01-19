package services

import (
	"bytes"
	"fmt"

	"text/template"

	"checkin-system/config"
	"checkin-system/models"

	"gopkg.in/gomail.v2"
)

// EmailService 邮件服务
type EmailService struct {
	config    config.EmailConfig
	dialer    *gomail.Dialer
	templates config.EmailTemplates
}

// NewEmailService 创建邮件服务实例
func NewEmailService(emailConfig config.EmailConfig) *EmailService {
	dialer := gomail.NewDialer(
		emailConfig.SMTPHost,
		587,
		emailConfig.SMTPEmail,
		emailConfig.SMTPPassword,
	)

	templates, err := config.LoadEmailTemplates()
	if err != nil {
		fmt.Printf("Warning: Failed to load email templates: %v\n", err)
		templates = make(config.EmailTemplates)
	}

	return &EmailService{
		config:    emailConfig,
		dialer:    dialer,
		templates: templates,
	}
}

// SendWelcomeEmail 发送欢迎邮件
func (e *EmailService) SendWelcomeEmail(user *models.User) error {
	template, exists := e.templates["welcome"]
	if !exists {
		return fmt.Errorf("welcome email template not found")
	}

	subject, body, err := e.parseTemplate(template, map[string]interface{}{
		"Username": user.Username,
		"Email":    user.Email,
	})
	if err != nil {
		return err
	}

	return e.sendEmail(user.Email, subject, body)
}

// SendDailyReminder 发送每日提醒邮件
func (e *EmailService) SendDailyReminder(user *models.User) error {
	template, exists := e.templates["daily_reminder"]
	if !exists {
		return fmt.Errorf("daily reminder email template not found")
	}

	subject, body, err := e.parseTemplate(template, map[string]interface{}{
		"Username": user.Username,
	})
	if err != nil {
		return err
	}

	return e.sendEmail(user.Email, subject, body)
}

// SendHourlyReminder 发送小时提醒邮件
func (e *EmailService) SendHourlyReminder(user *models.User) error {
	template, exists := e.templates["hourly_reminder"]
	if !exists {
		return fmt.Errorf("hourly reminder email template not found")
	}

	subject, body, err := e.parseTemplate(template, map[string]interface{}{
		"Username": user.Username,
	})
	if err != nil {
		return err
	}

	return e.sendEmail(user.Email, subject, body)
}

// SendMissedCheckInWarning 发送缺签警告邮件
func (e *EmailService) SendMissedCheckInWarning(user *models.User) error {
	template, exists := e.templates["missed_checkin_warning"]
	if !exists {
		return fmt.Errorf("missed checkin warning email template not found")
	}

	subject, body, err := e.parseTemplate(template, map[string]interface{}{
		"Username": user.Username,
	})
	if err != nil {
		return err
	}

	return e.sendEmail(user.Email, subject, body)
}

// SendTestEmail 发送测试邮件
func (e *EmailService) SendTestEmail(user *models.User) error {
	template, exists := e.templates["test_email"]
	if !exists {
		return fmt.Errorf("test email template not found")
	}

	subject, body, err := e.parseTemplate(template, map[string]interface{}{
		"Username": user.Username,
	})
	if err != nil {
		return err
	}

	return e.sendEmail(user.Email, subject, body)
}

// SendEmailVerification 发送邮箱验证邮件
func (e *EmailService) SendEmailVerification(user *models.User, verificationURL string) error {
	template, exists := e.templates["email_verification"]
	if !exists {
		return fmt.Errorf("email verification template not found")
	}

	subject, body, err := e.parseTemplate(template, map[string]interface{}{
		"Username":        user.Username,
		"VerificationURL": verificationURL,
	})
	if err != nil {
		return err
	}

	return e.sendEmail(user.Email, subject, body)
}

// parseTemplate 解析邮件模板
func (e *EmailService) parseTemplate(emailTemplate config.EmailTemplate, data map[string]interface{}) (subject, body string, err error) {
	// 解析主题
	subjectTmpl, err := template.New("subject").Parse(emailTemplate.Subject)
	if err != nil {
		return "", "", err
	}

	var subjectBuf bytes.Buffer
	if err := subjectTmpl.Execute(&subjectBuf, data); err != nil {
		return "", "", err
	}

	// 解析正文
	bodyTmpl, err := template.New("body").Parse(emailTemplate.Body)
	if err != nil {
		return "", "", err
	}

	var bodyBuf bytes.Buffer
	if err := bodyTmpl.Execute(&bodyBuf, data); err != nil {
		return "", "", err
	}

	return subjectBuf.String(), bodyBuf.String(), nil
}

// sendEmail 发送邮件
func (e *EmailService) sendEmail(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", e.config.SMTPEmail)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	return e.dialer.DialAndSend(m)
}

// ReloadTemplates 重新加载邮件模板
func (e *EmailService) ReloadTemplates() error {
	templates, err := config.LoadEmailTemplates()
	if err != nil {
		return err
	}
	e.templates = templates
	return nil
}
