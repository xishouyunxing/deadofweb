package handlers

import (
	"net/http"
	"time"

	"checkin-system/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ReminderHandler 提醒处理器
type ReminderHandler struct {
	db *gorm.DB
}

// NewReminderHandler 创建提醒处理器
func NewReminderHandler(db *gorm.DB) *ReminderHandler {
	return &ReminderHandler{
		db: db,
	}
}

// GetReminder 获取提醒设置
func (h *ReminderHandler) GetReminder(c *gin.Context) {
	userID := c.GetUint("user_id")

	var reminder models.CheckInReminder
	if err := h.db.Where("user_id = ?", userID).First(&reminder).Error; err != nil {
		// 如果没有找到，创建默认设置
		reminder = models.CheckInReminder{
			UserID:           userID,
			IsEnabled:        true,
			ReminderFrequency: "daily",
			ReminderInterval: 24,
			NextReminder:     time.Now().AddDate(0, 0, 1),
		}
		h.db.Create(&reminder)
	}

	c.JSON(http.StatusOK, reminder)
}

// UpdateReminderRequest 更新提醒设置请求
type UpdateReminderRequest struct {
	IsEnabled         *bool   `json:"is_enabled"`
	ReminderFrequency *string `json:"reminder_frequency"`
	ReminderInterval  *int    `json:"reminder_interval"`
}

// UpdateReminder 更新提醒设置
func (h *ReminderHandler) UpdateReminder(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req UpdateReminderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var reminder models.CheckInReminder
	if err := h.db.Where("user_id = ?", userID).First(&reminder).Error; err != nil {
		// 如果没有找到，创建新的
		reminder = models.CheckInReminder{
			UserID: userID,
		}
	}

	// 更新字段
	if req.IsEnabled != nil {
		reminder.IsEnabled = *req.IsEnabled
	}

	if req.ReminderFrequency != nil {
		reminder.ReminderFrequency = *req.ReminderFrequency
	}

	if req.ReminderInterval != nil {
		reminder.ReminderInterval = *req.ReminderInterval
	}

	// 如果启用了提醒，重新计算下次提醒时间
	if reminder.IsEnabled {
		var lastCheckIn time.Time
		h.db.Model(&models.CheckIn{}).
			Where("user_id = ?", userID).
			Order("checkin_at DESC").
			Limit(1).
			Pluck("checkin_at", &lastCheckIn)
		
		reminder.NextReminder = reminder.CalculateNextReminder(lastCheckIn)
	} else {
		reminder.NextReminder = time.Time{}
	}

	if err := h.db.Save(&reminder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update reminder settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Reminder settings updated successfully",
		"reminder": reminder,
	})
}