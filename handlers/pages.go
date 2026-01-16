package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// IndexHandler 首页处理器
func IndexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "签到系统",
	})
}

// LoginPageHandler 登录页面处理器
func LoginPageHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "登录",
	})
}

// RegisterPageHandler 注册页面处理器
func RegisterPageHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", gin.H{
		"title": "注册",
	})
}

// DashboardHandler 仪表板页面处理器
func DashboardHandler(c *gin.Context) {
	userID := c.GetUint("user_id")
	username := c.GetString("username")

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title":    "仪表板",
		"user_id":  userID,
		"username": username,
	})
}