package main

import (
	"fmt"
	"go-ledger/config"
	"go-ledger/routers"
)

func main() {
	// 先加载配置文件
	config.InitConfig()
	// 再初始化数据库连接
	config.InitDB()
	r := routers.SetupRouter()
	r.Run(":8080")
	fmt.Println("Hello, World!")
}
