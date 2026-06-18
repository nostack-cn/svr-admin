package router

import (
	"github.com/gin-gonic/gin"
	"github.com/nostack-cn/svr-admin/handler"
	"github.com/nostack-cn/svr-admin/middleware"
)

// Setup 注册所有路由
func Setup(r *gin.Engine) {
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())

	// 健康检查
	r.GET("/health", handler.HealthCheck)

	// API v1 分组
	v1 := r.Group("/api/v1")
	{
		_ = v1 // 在此注册业务路由
	}
}
