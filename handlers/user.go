package handlers

import (
	"net/http"

	"checkin-system/middleware"
	"checkin-system/models"
	"checkin-system/services"

	"github.com/gin-contrib/sessions"
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

// Register 用户注册
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查用户名是否已存在
	var existingUser models.User
	if err := h.db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username or email already exists"})
		return
	}

	// 创建用户
	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// 创建默认签到提醒设置
	reminder := models.CheckInReminder{
		UserID:           user.ID,
		IsEnabled:        true,
		ReminderFrequency: "daily",
		ReminderInterval: 24,
		NextReminder:     time.Now().AddDate(0, 0, 1), // 明天提醒
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

	// 自动登录，设置session
	middleware.SetSession(c, user.ID, user.Username)

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user":    user.ToSafeUser(),
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