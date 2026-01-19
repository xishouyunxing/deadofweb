package models

import (
	"time"
)

type CheckInReminder struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	UserID            uint      `json:"user_id" gorm:"not null;uniqueIndex"`
	IsEnabled         bool      `json:"is_enabled" gorm:"default:true"`
	ReminderFrequency string    `json:"reminder_frequency" gorm:"default:'daily'"` // daily, hourly, custom
	ReminderInterval  int       `json:"reminder_interval" gorm:"default:24"`       // 小时数
	NextReminder      time.Time `json:"next_reminder"`
	LastReminder      time.Time `json:"last_reminder"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// CalculateNextReminder 计算下次提醒时间
func (r *CheckInReminder) CalculateNextReminder(lastCheckIn time.Time) time.Time {
	if !r.IsEnabled {
		return time.Time{}
	}

	switch r.ReminderFrequency {
	case "daily":
		// 每天提醒一次
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())
		if next.Before(now) {
			next = next.AddDate(0, 0, 1)
		}
		return next
	case "hourly":
		// 按小时间隔提醒
		// 优先使用最后一次提醒时间作为基准，如果没有则使用最后签到时间，如果都没有则使用当前时间
		baseTime := time.Now()
		if !r.LastReminder.IsZero() {
			baseTime = r.LastReminder
		} else if !lastCheckIn.IsZero() {
			baseTime = lastCheckIn
		}
		return baseTime.Add(time.Duration(r.ReminderInterval) * time.Hour)
	default:
		// 自定义间隔
		return time.Now().Add(time.Duration(r.ReminderInterval) * time.Hour)
	}
}
