package database

import (
	"log"

	"checkin-system/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() *gorm.DB {
	dbConfig := config.GetDatabaseConfig()
	
	dsn := dbConfig.GetDSN()
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	
	// 测试连接
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get database instance:", err)
	}
	
	if err := sqlDB.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	
	log.Println("Database connected successfully")
	DB = db
	return db
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}