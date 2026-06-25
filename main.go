package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/nostack-cn/svr-admin/config"
	"github.com/nostack-cn/svr-admin/cron"
	"github.com/nostack-cn/svr-admin/handler"
	"github.com/nostack-cn/svr-admin/model"
	"github.com/nostack-cn/svr-admin/pkg/auth"
	"github.com/nostack-cn/svr-admin/pkg/consoleclient"
	"github.com/nostack-cn/svr-admin/pkg/profile"
	"github.com/nostack-cn/svr-admin/router"
	"github.com/nostack-cn/svr-admin/service"
)

func main() {
	// 1. 解析命令行参数
	config.ParseFlags()

	// 2. 加载配置
	cfg := config.Load()
	if configFile := config.ConfigFile(); configFile != "" {
		log.Printf("[Config] 使用配置文件: %s", configFile)
	} else {
		log.Println("[Config] 未使用配置文件，采用环境变量 + 默认值")
	}

	// 3. 初始化数据库
	db, err := gorm.Open(mysql.Open(cfg.Database.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("[DB] 连接失败: %v", err)
	}
	if err := autoMigrate(db); err != nil {
		log.Fatalf("[DB] 迁移失败: %v", err)
	}

	// 4. 初始化依赖
	jwtMgr := auth.NewJWTManager(cfg.JWT.Secret)
	profileClient := initProfileClient(cfg)
	consoleClient := initConsoleClient(cfg)

	rbacService := service.NewRBACService(db)
	adminService := service.NewAdminService(db, jwtMgr, rbacService)
	auditService := service.NewAuditService(db)
	userAdminService := service.NewUserAdminService(profileClient)
	orderAdminService := service.NewOrderAdminService(profileClient, consoleClient)

	// 5. 初始化 RBAC 种子数据与初始超管
	if err := seed(cfg, rbacService, adminService); err != nil {
		log.Fatalf("[Seed] 初始化失败: %v", err)
	}

	// 6. 组装 Handler
	handlers := &router.Handlers{
		Auth:  handler.NewAuthHandler(adminService),
		Admin: handler.NewAdminHandler(adminService),
		RBAC:  handler.NewRBACHandler(rbacService),
		User:  handler.NewUserHandler(userAdminService),
		Order: handler.NewOrderHandler(orderAdminService),
		Log:   handler.NewLogHandler(auditService),
	}

	// 7. 定时任务
	cron.Setup()

	// 8. Gin 引擎
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	router.Setup(r, handlers, jwtMgr, auditService)

	// 9. 启动
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("[Server] svr-admin 启动于 %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("[Server] 启动失败: %v", err)
	}
}

// autoMigrate 自动迁移所有模型
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Permission{},
		&model.Role{},
		&model.Admin{},
		&model.AdminOperationLog{},
	)
}

// initProfileClient 初始化 svr-profile 下游客户端
func initProfileClient(cfg *config.Config) *profile.Client {
	if cfg.Profile.BaseURL == "" {
		log.Println("[Profile] svr-profile 未配置（PROFILE_BASE_URL 为空），用户/订单管控功能不可用")
		return nil
	}
	log.Printf("[Profile] svr-profile 客户端初始化完成 (BaseURL: %s)", cfg.Profile.BaseURL)
	return profile.NewClient(cfg.Profile.BaseURL, cfg.Profile.InternalKey)
}

// initConsoleClient 初始化 svr-console 下游客户端
func initConsoleClient(cfg *config.Config) *consoleclient.Client {
	if cfg.Console.BaseURL == "" {
		log.Println("[Console] svr-console 未配置（CONSOLE_BASE_URL 为空），订阅取消功能不可用")
		return nil
	}
	log.Printf("[Console] svr-console 客户端初始化完成 (BaseURL: %s)", cfg.Console.BaseURL)
	return consoleclient.NewClient(cfg.Console.BaseURL, cfg.Console.InternalKey)
}

// seed 初始化权限点、超级管理员角色及初始超管账号
func seed(cfg *config.Config, rbac *service.RBACService, adminSvc *service.AdminService) error {
	if err := rbac.SeedPermissions(); err != nil {
		return fmt.Errorf("同步权限点失败: %w", err)
	}
	roleID, err := rbac.EnsureSuperAdminRole()
	if err != nil {
		return fmt.Errorf("初始化超管角色失败: %w", err)
	}
	created, err := adminSvc.SeedSuperAdmin(
		cfg.Seed.SuperAdminUsername,
		cfg.Seed.SuperAdminPassword,
		cfg.Seed.SuperAdminEmail,
		roleID,
	)
	if err != nil {
		return fmt.Errorf("创建初始超管失败: %w", err)
	}
	if created {
		log.Printf("[Seed] 已创建初始超级管理员: %s", cfg.Seed.SuperAdminUsername)
	}
	return nil
}
