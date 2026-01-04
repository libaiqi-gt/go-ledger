package services

import (
	"errors"
	"go-ledger/config"
	"go-ledger/models"
	"go-ledger/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct{}

// Register 用户注册业务逻辑
func (s *AuthService) Register(username, password string) (*models.User, error) {
	// 1. 检查用户名是否存在
	var existingUser models.User
	if err := config.DB.Where("username = ?", username).First(&existingUser).Error; err == nil {
		return nil, errors.New("用户名已存在")
	} else if err != gorm.ErrRecordNotFound {
		return nil, err // 数据库错误
	}

	// 2. 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	// 3. 创建用户
	user := models.User{
		Username: username,
		Password: string(hashedPassword),
	}

	if err := config.DB.Create(&user).Error; err != nil {
		return nil, err
	}

	return &user, nil
}

// Login 用户登录业务逻辑
func (s *AuthService) Login(username, password string) (string, error) {
	// 1. 查找用户
	var user models.User
	if err := config.DB.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", errors.New("用户名或密码错误")
		}
		return "", err
	}

	// 2. 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("用户名或密码错误")
	}

	// 3. 生成 Token
	token, err := utils.GenerateToken(user.ID)
	if err != nil {
		return "", errors.New("Token生成失败")
	}

	return token, nil
}
