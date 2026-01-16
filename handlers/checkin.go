package handlers

import (
	"fmt"
	"net/http"
	"time"

	"checkin-system/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CheckInHandler 签到处理器
type CheckInHandler struct {
	db *gorm.DB
}

// NewCheckInHandler 创建签到处理器
func NewCheckInHandler(db *gorm.DB) *CheckInHandler {
	return &CheckInHandler{
		db: db,
	}
}

// CheckInRequest 签到请求
type CheckInRequest struct {
	Note string `json:"note"`
}

// CheckIn 用户签到
func (h *CheckInHandler) CheckIn(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req CheckInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 检查今天是否已经签到
	var todayCheckIn models.CheckIn
	today := time.Now()
	err := h.db.Where("user_id = ? AND DATE(checkin_at) = CURRENT_DATE", userID).First(&todayCheckIn).Error
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Already checked in today"})
		return
	}

	// 创建签到记录
	checkIn := models.CheckIn{
		UserID:    userID,
		CheckInAt: time.Now(),
		Note:      req.Note,
	}

	if err := h.db.Create(&checkIn).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check in"})
		return
	}

	// 更新下次提醒时间
	var reminder models.CheckInReminder
	if err := h.db.Where("user_id = ?", userID).First(&reminder).Error; err == nil {
		reminder.NextReminder = reminder.CalculateNextReminder(checkIn.CheckInAt)
		h.db.Save(&reminder)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Check in successful",
		"checkin": checkIn,
	})
}

// GetCheckInHistory 获取签到历史
func (h *CheckInHandler) GetCheckInHistory(c *gin.Context) {
	userID := c.GetUint("user_id")

	// 分页参数
	page := 1
	if p := c.Query("page"); p != "" {
		if parsed, err := parseInt(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsed, err := parseInt(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	offset := (page - 1) * limit

	var checkIns []models.CheckIn
	var total int64

	// 获取总数
	h.db.Model(&models.CheckIn{}).Where("user_id = ?", userID).Count(&total)

	// 获取分页数据
	if err := h.db.Where("user_id = ?", userID).
		Order("checkin_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&checkIns).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch check-in history"})
		return
	}

	// 计算连续签到天数
	consecutiveDays := models.GetConsecutiveDays(checkIns)

	c.JSON(http.StatusOK, gin.H{
		"checkins":          checkIns,
		"total":             total,
		"page":              page,
		"limit":             limit,
		"consecutive_days":  consecutiveDays,
	})
}

// GetCheckInStatus 获取签到状态
func (h *CheckInHandler) GetCheckInStatus(c *gin.Context) {
	userID := c.GetUint("user_id")

	// 检查今天是否已签到
	var todayCheckIn models.CheckIn
	err := h.db.Where("user_id = ? AND DATE(checkin_at) = CURRENT_DATE", userID).First(&todayCheckIn).Error
	
	todayChecked := err == nil

	// 获取最近7天的签到记录
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	var recentCheckIns []models.CheckIn
	
	h.db.Where("user_id = ? AND checkin_at >= ?", userID, sevenDaysAgo).
		Order("checkin_at DESC").
		Find(&recentCheckIns)

	// 计算连续签到天数
	consecutiveDays := models.GetConsecutiveDays(recentCheckIns)

	// 获取本月签到次数
	monthStart := time.Date(time.Now().Year(), time.Now().Month(), 1, 0, 0, 0, 0, time.Now().Location())
	var monthCount int64
	h.db.Model(&models.CheckIn{}).
		Where("user_id = ? AND checkin_at >= ?", userID, monthStart).
		Count(&monthCount)

	status := gin.H{
		"today_checked":    todayChecked,
		"consecutive_days":  consecutiveDays,
		"month_count":       monthCount,
		"recent_checkins":   recentCheckIns,
	}

	if todayChecked {
		status["last_checkin"] = todayCheckIn
	}

	c.JSON(http.StatusOK, status)
}

// parseInt 辅助函数：将字符串转换为整数
func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}