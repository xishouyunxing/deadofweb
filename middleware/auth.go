package middleware

import (
	"net/http"
	"checkin-system/database"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthMiddleware Session认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		
		// 检查session中是否有用户ID
		userID := session.Get("user_id")
		if userID == nil {
			if isAPIRequest(c) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			} else {
				c.Redirect(http.StatusFound, "/login")
			}
			c.Abort()
			return
		}

		// 验证用户是否存在
		var user struct {
			ID       uint   `json:"id"`
			Username string `json:"username"`
		}
		
		db := database.GetDB()
		if err := db.Model(&struct {
			ID       uint   `json:"id"`
			Username string `json:"username"`
		}{}).Table("users").Select("id, username").Where("id = ?", userID).First(&user).Error; err != nil {
			if isAPIRequest(c) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
			} else {
				// 清除无效session并重定向
				session.Clear()
				session.Save()
				c.Redirect(http.StatusFound, "/login")
			}
			c.Abort()
			return
		}

		// 设置用户信息到context
		c.Set("user_id", user.ID)
		c.Set("username", user.Username)
		c.Set("authenticated", true)

		c.Next()
	}
}

// OptionalAuthMiddleware 可选认证中间件
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		
		if userID != nil {
			db := database.GetDB()
			var user struct {
				ID       uint   `json:"id"`
				Username string `json:"username"`
			}
			
			if err := db.Model(&struct {
				ID       uint   `json:"id"`
				Username string `json:"username"`
			}{}).Table("users").Select("id, username").Where("id = ?", userID).First(&user).Error; err == nil {
				c.Set("user_id", user.ID)
				c.Set("username", user.Username)
				c.Set("authenticated", true)
			} else {
				// 用户不存在，清除session
				session.Clear()
				session.Save()
				c.Set("authenticated", false)
			}
		} else {
			c.Set("authenticated", false)
		}
		
		c.Next()
	}
}

// isAPIRequest 检查是否为API请求
func isAPIRequest(c *gin.Context) bool {
	return len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api"
}

// SetSession 设置用户session
func SetSession(c *gin.Context, userID uint, username string) {
	session := sessions.Default(c)
	session.Set("user_id", userID)
	session.Set("username", username)
	session.Save()
}

// ClearSession 清除用户session
func ClearSession(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
}