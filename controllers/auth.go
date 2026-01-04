package controllers

import (
	"go-ledger/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 定义接收前端参数的结构体
type RegisterInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

var authService = new(services.AuthService)

// Register 注册接口
func Register(c *gin.Context) {
	var input RegisterInput

	// 1. 参数校验
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. 调用 Service 注册
	user, err := authService.Register(input.Username, input.Password)
	if err != nil {
		// 这里简单处理，具体错误码可以根据 err 类型细分
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "注册成功", "data": user.Username})
}

// Login 登录接口
func Login(c *gin.Context) {
	var input RegisterInput // 复用上面的结构体，字段一样
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. 调用 Service 登录
	token, err := authService.Login(input.Username, input.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// 2. 返回 Token
	c.JSON(http.StatusOK, gin.H{"token": token})
}
