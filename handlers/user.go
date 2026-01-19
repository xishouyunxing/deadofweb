package handlers

import (
	"checkin-system/middleware"
	"checkin-system/models"
	"checkin-system/services"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	//"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UserHandler 用户处理器
type UserHandler struct {
	db           *gorm.DB
	emailService *services.EmailService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(db *gorm.DB, emailService *services.EmailService) *UserHandler {
	return &UserHandler{
		db:           db,
		emailService: emailService,
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// generateVerificationToken 生成验证令牌
func generateVerificationToken() (string, error) {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Register 用户注册
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查用户名或邮箱是否已存在（包括软删除的记录）
	var existingUser models.User
	if err := h.db.Unscoped().Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
		// 如果存在已软删除的记录，清理相关数据并硬删除用户
		if existingUser.DeletedAt.Valid {
			// 使用事务确保数据一致性
			tx := h.db.Unscoped().Begin()
			defer func() {
				if r := recover(); r != nil {
					tx.Rollback()
				}
			}()

			// 删除用户的所有签到记录
			if err := tx.Delete(&models.CheckIn{}, "user_id = ?", existingUser.ID).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clean up existing checkin records"})
				return
			}

			// 删除用户的提醒设置
			if err := tx.Delete(&models.CheckInReminder{}, "user_id = ?", existingUser.ID).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clean up existing reminder settings"})
				return
			}

			// 硬删除用户本身
			if err := tx.Delete(&existingUser).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clean up existing user data"})
				return
			}

			// 提交事务
			if err := tx.Commit().Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete data cleanup"})
				return
			}
		} else {
			// 如果存在未删除的记录，返回冲突错误
			c.JSON(http.StatusConflict, gin.H{"error": "Username or email already exists"})
			return
		}
	}

	// 生成验证令牌
	verificationToken, err := generateVerificationToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate verification token"})
		return
	}

	// 创建用户
	user := models.User{
		Username:                   req.Username,
		Email:                      req.Email,
		Password:                   req.Password,
		EmailVerified:              false,
		VerificationToken:          verificationToken,
		VerificationTokenExpiresAt: time.Now().Add(24 * time.Hour), // 令牌24小时内有效
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// 创建默认签到提醒设置
	reminder := models.CheckInReminder{
		UserID:            user.ID,
		IsEnabled:         true,
		ReminderFrequency: "daily",
		ReminderInterval:  24,
		NextReminder:      time.Now().AddDate(0, 0, 1), // 明天提醒
	}

	if err := h.db.Create(&reminder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reminder settings"})
		return
	}

	// 发送欢迎邮件
	go func() {
		if err := h.emailService.SendWelcomeEmail(&user); err != nil {
			// 记录错误但不影响注册流程
			// 可以添加日志记录
		}
	}()

	c.JSON(http.StatusCreated, gin.H{
		"message":        "User registered successfully. You can now login.",
		"user":           user.ToSafeUser(),
		"email_verified": false,
	})
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查找用户
	var user models.User
	if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// 验证密码
	if !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// 设置session
	middleware.SetSession(c, user.ID, user.Username)

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user":    user.ToSafeUser(),
	})
}

// GetProfile 获取用户信息
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user.ToSafeUser(),
	})
}

// UpdateProfileRequest 更新用户信息请求
type UpdateProfileRequest struct {
	Email string `json:"email" binding:"omitempty,email"`
}

// UpdateProfile 更新用户信息
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// 更新邮箱（如果提供）
	if req.Email != "" {
		user.Email = req.Email
	}

	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user":    user.ToSafeUser(),
	})
}

// Logout 用户登出
func (h *UserHandler) Logout(c *gin.Context) {
	middleware.ClearSession(c)
	c.JSON(http.StatusOK, gin.H{
		"message": "Logout successful",
	})
}

// Cancel 用户注销
func (h *UserHandler) Cancel(c *gin.Context) {
	// 获取当前登录用户ID
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	// 使用事务确保数据一致性
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 删除用户的所有签到记录
	if err := tx.Where("user_id = ?", userID).Delete(&models.CheckIn{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete checkin records"})
		return
	}

	// 删除用户的提醒设置
	if err := tx.Where("user_id = ?", userID).Delete(&models.CheckInReminder{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete reminder settings"})
		return
	}

	// 删除用户本身
	if err := tx.Delete(&models.User{}, userID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user account"})
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete cancellation"})
		return
	}

	// 清除session
	middleware.ClearSession(c)

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"message": "User account cancelled successfully",
	})
}

// SendTestEmail 发送测试邮件
func (h *UserHandler) SendTestEmail(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 发送测试邮件
	go func() {
		if err := h.emailService.SendTestEmail(&user); err != nil {
			// 记录错误但不影响响应
			// 可以添加日志记录
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "测试邮件已发送，请检查您的邮箱",
	})
}

// SendVerificationEmail 发送验证邮件
func (h *UserHandler) SendVerificationEmail(c *gin.Context) {
	userID := c.GetUint("user_id")
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}

	// 查找用户
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 如果邮箱已验证，返回错误
	if user.EmailVerified {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email already verified"})
		return
	}

	// 生成新的验证令牌
	verificationToken, err := generateVerificationToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate verification token"})
		return
	}

	// 更新用户的验证令牌信息
	user.VerificationToken = verificationToken
	user.VerificationTokenExpiresAt = time.Now().Add(24 * time.Hour) // 令牌24小时内有效

	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update verification token"})
		return
	}

	// 发送验证邮件
	verificationURL := c.Request.Host + "/api/verify-email?token=" + verificationToken
	go func() {
		if err := h.emailService.SendEmailVerification(&user, verificationURL); err != nil {
			// 记录错误但不影响响应
			// 可以添加日志记录
		}
	}()

	c.JSON(http.StatusOK, gin.H{
		"message": "Verification email sent. Please check your inbox and spam folder.",
	})
}

// VerifyEmail 验证邮箱
func (h *UserHandler) VerifyEmail(c *gin.Context) {
	// 获取验证令牌
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Verification token is required"})
		return
	}

	// 查找用户
	var user models.User
	if err := h.db.Where("verification_token = ?", token).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired verification token"})
		return
	}

	// 检查令牌是否过期
	if time.Now().After(user.VerificationTokenExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Verification token has expired"})
		return
	}

	// 更新用户信息
	user.EmailVerified = true
	user.VerificationToken = ""
	user.VerificationTokenExpiresAt = time.Time{}

	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify email"})
		return
	}

	// 重定向到登录页面或返回成功信息
	c.JSON(http.StatusOK, gin.H{
		"message": "Email verified successfully. You can now login.",
	})
}
