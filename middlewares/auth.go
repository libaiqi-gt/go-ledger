package middlewares

import (
	"go-ledger/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func JwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取请求头中的 Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			c.Abort() // 阻止后续处理
			return
		}

		// 2. 格式校验：通常是 "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization format must be Bearer <token>"})
			c.Abort()
			return
		}

		// 3. 解析 Token
		tokenString := parts[1]
		userID, err := utils.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// 4. 将 userID 存入上下文，供后续 Controller 使用
		// 注意：这里存进去的是 float64 (JWT默认)，取的时候要注意转换，或者在 ParseToken 里转成 uint
		c.Set("userID", uint(userID))

		// 5. 放行，进入下一个 Handler
		c.Next()
	}
}
