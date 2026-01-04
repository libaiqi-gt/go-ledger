package main

import (
	"fmt"
	"go-ledger/config"
	"go-ledger/routers"
	"os"
)

func main() {
	// 1. 读取 Zeabur 的 PORT 环境变量，兜底默认端口 3000
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // 本地开发时的默认端口
	}
	// 先加载配置文件
	config.InitConfig()
	// 再初始化数据库连接
	config.InitDB()
	r := routers.SetupRouter()
	listenAddr := fmt.Sprintf("0.0.0.0:%s", port)
	fmt.Printf("服务正在监听地址：%s\n", listenAddr)
	err := r.Run(listenAddr) // Gin 的 Run 方法封装了 http.ListenAndServe
	if err != nil {
		fmt.Printf("服务启动失败：%v\n", err)
	}
}
