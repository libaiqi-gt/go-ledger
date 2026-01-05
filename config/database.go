package config

import (
	"fmt"
	"go-ledger/models"
	"os"
	"strings"

	"github.com/joho/godotenv"
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

	// 关键：必须显式调用 godotenv 加载 .env 文件到系统环境变量
	// 否则 viper 读取到的可能是空值，或者 ${ENV_VAR} 字符串本身
	_ = godotenv.Load()

	// 重新读取配置以解析 ${ENV_VAR}
	// viper 默认不会自动替换 config.yaml 中的 ${VAR}，需要手动处理或使用 ExpandEnv
	// 这里我们直接依赖 viper.AutomaticEnv() 和 viper.GetString() 的结合
	// 但要注意：viper.GetString("ai.api_key") 此时可能直接返回 "${AI_API_KEY}" 字符串
	// 解决方法：使用 os.ExpandEnv 处理读取到的值，或者让 viper 自动处理
	// 更好的方式：既然已经在 config.yaml 里写了 ${VAR}，我们可以在读取时手动 Expand
}

// InitDB 初始化数据库连接（读取配置并建立连接）
func InitDB() {
	host := os.Getenv("MYSQL_HOST")
	if host == "" {
		host = viper.GetString("database.host")
	}
	port := os.Getenv("MYSQL_PORT")
	if port == "" {
		port = viper.GetString("database.port")
	}
	user := os.Getenv("MYSQL_USERNAME")
	if user == "" {
		user = viper.GetString("database.user")
	}
	password := os.Getenv("MYSQL_PASSWORD")
	if password == "" {
		password = viper.GetString("database.password")
	}
	dbname := os.Getenv("MYSQL_DATABASE")
	if dbname == "" {
		dbname = viper.GetString("database.dbname")
	}

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
