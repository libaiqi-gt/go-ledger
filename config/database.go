package config

import (
	"fmt"
	"go-ledger/models"
	"strings"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitConfig 使用 viper 加载配置文件（config/config.yaml）
func InitConfig() {
	// 指定配置文件路径与类型
	viper.SetConfigFile("config/config.yaml")
	viper.SetConfigType("yaml")
	// 读取配置文件，失败直接退出（也可按需返回错误）
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Warning: 读取配置文件失败: %v\n", err)
	}

	// 开启环境变量支持
	// 将 . 替换为 _，例如 database.host -> DATABASE_HOST
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}

// InitDB 初始化数据库连接（读取配置并建立连接）
func InitDB() {
	host := viper.GetString("database.host")
	port := viper.GetString("database.port")
	user := viper.GetString("database.user")
	password := viper.GetString("database.password")
	dbname := viper.GetString("database.dbname")

	dns := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, dbname)
	database, err := gorm.Open(mysql.Open(dns), &gorm.Config{})
	if err != nil {
		panic("连接数据库失败")
	}
	// 自动迁移模式，自动创建数据库
	database.AutoMigrate(&models.User{}, &models.LedgerEntry{})

	DB = database
	fmt.Println("数据库连接成功")
}
