package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

// 定义 JWT 的密钥 (生产环境请放到环境变量中，不要硬编码)
var jwtKey = []byte(viper.GetString("jwt.secret"))

// GenerateToken 生成 JWT Token
func GenerateToken(userID uint) (string, error) {
	// 定义 Claims (载荷)
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // 过期时间：24小时
	}

	// 生成 Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

// 解析并验证token
func ParseToken(tokenString string) (float64, error) {
	// 解析token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// 确保使用的是 HMAC 签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtKey, nil
	})
	if err != nil {
		return 0, err
	}
	// 验证token是否有效，并提取Claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := claims["user_id"].(float64)
		return userID, nil
	} else {
		return 0, fmt.Errorf("invalid token")
	}
}
