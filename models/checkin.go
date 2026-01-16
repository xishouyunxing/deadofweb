package models

import (
	"time"
)

type CheckIn struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	CheckInAt time.Time `json:"checkin_at" gorm:"not null"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

// IsToday 检查是否是今天的签到
func (c *CheckIn) IsToday() bool {
	now := time.Now()
	return c.CheckInAt.Year() == now.Year() &&
		c.CheckInAt.Month() == now.Month() &&
		c.CheckInAt.Day() == now.Day()
}

// GetConsecutiveDays 获取连续签到天数
func GetConsecutiveDays(checkIns []CheckIn) int {
	if len(checkIns) == 0 {
		return 0
	}

	consecutive := 0
	currentDate := time.Now()
	
	for i := len(checkIns) - 1; i >= 0; i-- {
		checkInDate := checkIns[i].CheckInAt
		
		// 检查是否是连续的日期
		if checkInDate.Year() == currentDate.Year() &&
			checkInDate.Month() == currentDate.Month() &&
			checkInDate.Day() == currentDate.Day() {
			
			consecutive++
			currentDate = currentDate.AddDate(0, 0, -1)
		} else if checkInDate.After(currentDate) {
			// 如果签到日期晚于当前检查日期，跳过
			continue
		} else {
			// 如果不是连续的，中断
			break
		}
	}
	
	return consecutive
}