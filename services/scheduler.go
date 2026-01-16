package services

import (
	"log"
	"time"

	"checkin-system/models"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

// SchedulerService 定时任务服务
type SchedulerService struct {
	db           *gorm.DB
	emailService *EmailService
	cron         *cron.Cron
}

// NewSchedulerService 创建定时任务服务
func NewSchedulerService(db *gorm.DB, emailService *EmailService) *SchedulerService {
	c := cron.New()
	
	return &SchedulerService{
		db:           db,
		emailService: emailService,
		cron:         c,
	}
}

// Start 启动定时任务
func (s *SchedulerService) Start() {
	// 每小时检查一次提醒任务
	s.cron.AddFunc("0 * * * *", s.checkReminders)
	
	// 每天早上8点检查缺签用户
	s.cron.AddFunc("0 8 * * *", s.checkMissedCheckIns)
	
	s.cron.Start()
	log.Println("Scheduler service started")
}

// Stop 停止定时任务
func (s *SchedulerService) Stop() {
	s.cron.Stop()
	log.Println("Scheduler service stopped")
}

// checkReminders 检查提醒任务
func (s *SchedulerService) checkReminders() {
	log.Println("Checking reminders...")
	
	var reminders []models.CheckInReminder
	err := s.db.Where("is_enabled = ?", true).Find(&reminders).Error
	if err != nil {
		log.Printf("Error fetching reminders: %v", err)
		return
	}
	
	now := time.Now()
	for _, reminder := range reminders {
		if reminder.NextReminder.Before(now) || reminder.NextReminder.Equal(now) {
			var user models.User
			if err := s.db.First(&user, reminder.UserID).Error; err != nil {
				log.Printf("Error finding user %d: %v", reminder.UserID, err)
				continue
			}
			
			// 发送提醒邮件
			var err error
			switch reminder.ReminderFrequency {
			case "daily":
				err = s.emailService.SendDailyReminder(&user)
			case "hourly":
				err = s.emailService.SendHourlyReminder(&user)
			default:
				err = s.emailService.SendHourlyReminder(&user)
			}
			
			if err != nil {
				log.Printf("Error sending reminder to user %d: %v", reminder.UserID, err)
				continue
			}
			
			// 更新下次提醒时间
			reminder.LastReminder = now
			var lastCheckIn time.Time
			s.db.Model(&models.CheckIn{}).
				Where("user_id = ?", reminder.UserID).
				Order("checkin_at DESC").
				Limit(1).
				Pluck("checkin_at", &lastCheckIn)
			
			reminder.NextReminder = reminder.CalculateNextReminder(lastCheckIn)
			
			if err := s.db.Save(&reminder).Error; err != nil {
				log.Printf("Error updating reminder %d: %v", reminder.ID, err)
			}
		}
	}
}

// checkMissedCheckIns 检查缺签用户
func (s *SchedulerService) checkMissedCheckIns() {
	log.Println("Checking missed check-ins...")
	
	// 获取所有用户
	var users []models.User
	if err := s.db.Find(&users).Error; err != nil {
		log.Printf("Error fetching users: %v", err)
		return
	}
	
	for _, user := range users {
		// 检查是否连续两天未签到
		if s.hasMissedCheckIns(user.ID) {
			// 发送缺签提醒邮件
			if err := s.emailService.SendMissedCheckInWarning(&user); err != nil {
				log.Printf("Error sending missed checkin warning to user %d: %v", user.ID, err)
			} else {
				log.Printf("Sent missed checkin warning to user %s", user.Username)
			}
		}
	}
}

// hasMissedCheckIns 检查用户是否连续两天未签到
func (s *SchedulerService) hasMissedCheckIns(userID uint) bool {
	// 获取最近3天的签到记录
	threeDaysAgo := time.Now().AddDate(0, 0, -3)
	var checkIns []models.CheckIn
	
	err := s.db.Where("user_id = ? AND checkin_at >= ?", userID, threeDaysAgo).
		Order("checkin_at DESC").
		Find(&checkIns).Error
	if err != nil {
		return false
	}
	
	// 检查今天和昨天是否有签到
	today := time.Now()
	yesterday := today.AddDate(0, 0, -1)
	
	hasToday := false
	hasYesterday := false
	
	for _, checkIn := range checkIns {
		if checkIn.CheckInAt.Year() == today.Year() &&
			checkIn.CheckInAt.Month() == today.Month() &&
			checkIn.CheckInAt.Day() == today.Day() {
			hasToday = true
		}
		
		if checkIn.CheckInAt.Year() == yesterday.Year() &&
			checkIn.CheckInAt.Month() == yesterday.Month() &&
			checkIn.CheckInAt.Day() == yesterday.Day() {
			hasYesterday = true
		}
	}
	
	// 如果今天和昨天都没有签到，则缺签
	return !hasToday && !hasYesterday
}