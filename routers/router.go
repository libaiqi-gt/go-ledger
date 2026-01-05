package routers

import (
	"go-ledger/controllers"
	"go-ledger/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 注册跨域中间件
	r.Use(middlewares.CorsMiddleware())

	api := r.Group("/v1")
	{
		// 公开路由（不需要登录）
		api.POST("/register", controllers.Register) // 用户注册
		api.POST("/login", controllers.Login)       // 用户登录
		auth := api.Group("/")
		auth.Use(middlewares.JwtAuthMiddleware()) // 挂载中间件
		{
			auth.POST("/entries", controllers.CreateEntry)           // 记账
			auth.POST("/entries/smart", controllers.CreateEntryByAI) // 智能记账 (AI)
			auth.GET("/entries", controllers.FindEntries)            // 分页查询账单
			auth.DELETE("/entries/:id", controllers.DeleteEntry)     // 删除账单
		}
	}
	return r
}
