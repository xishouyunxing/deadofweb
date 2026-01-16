package main

import (
	"log"
	"os"

	"checkin-system/config"
	"checkin-system/database"
	"checkin-system/handlers"
	"checkin-system/middleware"
	"checkin-system/models"
	"checkin-system/services"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 加载环境变量
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// 初始化数据库
	db := database.InitDB()

	// 自动迁移数据库表
	err := db.AutoMigrate(
		&models.User{},
		&models.CheckIn{},
		&models.CheckInReminder{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// 初始化服务
	emailService := services.NewEmailService(config.GetEmailConfig())
	schedulerService := services.NewSchedulerService(db, emailService)
	
	// 启动定时任务
	go schedulerService.Start()

	// 初始化路由
	r := gin.Default()
	
	// 配置Session存储
	store := cookie.NewStore([]byte(getEnv("SESSION_SECRET", "your-secret-key")))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7天
		Secure:   false,
		HttpOnly: true,
	})
	r.Use(sessions.Sessions("checkin-session", store))
	
	// 加载HTML模板
	r.LoadHTMLGlob("templates/*")
	
	// 静态文件
	r.Static("/static", "./static")

	// 初始化处理器
	userHandler := handlers.NewUserHandler(db, emailService)
	checkInHandler := handlers.NewCheckInHandler(db)
	reminderHandler := handlers.NewReminderHandler(db)

	// API路由组
	api := r.Group("/api")
	{
		// 用户相关
		api.POST("/register", userHandler.Register)
		api.POST("/login", userHandler.Login)
		api.POST("/logout", middleware.AuthMiddleware(), userHandler.Logout)
		api.GET("/profile", middleware.AuthMiddleware(), userHandler.GetProfile)
		api.PUT("/profile", middleware.AuthMiddleware(), userHandler.UpdateProfile)

		// 签到相关
		api.POST("/checkin", middleware.AuthMiddleware(), checkInHandler.CheckIn)
		api.GET("/checkin/history", middleware.AuthMiddleware(), checkInHandler.GetCheckInHistory)
		api.GET("/checkin/status", middleware.AuthMiddleware(), checkInHandler.GetCheckInStatus)

		// 提醒相关
		api.GET("/reminder", middleware.AuthMiddleware(), reminderHandler.GetReminder)
		api.PUT("/reminder", middleware.AuthMiddleware(), reminderHandler.UpdateReminder)
	}

	// 页面路由
	r.GET("/", handlers.IndexHandler)
	r.GET("/login", handlers.LoginPageHandler)
	r.GET("/register", handlers.RegisterPageHandler)
	r.GET("/dashboard", middleware.AuthMiddleware(), handlers.DashboardHandler)

	// 启动服务器
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}