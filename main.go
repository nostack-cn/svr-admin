package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/nostack-cn/svr-admin/config"
	"github.com/nostack-cn/svr-admin/cron"
	"github.com/nostack-cn/svr-admin/router"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 1. 加载配置
	cfg := config.Load()

	// 2. 初始化数据库
	db, err := gorm.Open(mysql.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("[DB] 连接失败: %v", err)
	}
	_ = db // 后续通过全局或依赖注入使用

	// 3. 初始化定时任务
	cron.Setup()

	// 4. 初始化 Gin 引擎
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	router.Setup(r)

	// 5. 启动服务
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("[Server] svr-admin 启动于 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("[Server] 启动失败: %v", err)
	}
}
