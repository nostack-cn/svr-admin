package router

import (
	"github.com/gin-gonic/gin"

	"github.com/nostack-cn/svr-admin/handler"
	"github.com/nostack-cn/svr-admin/middleware"
	"github.com/nostack-cn/svr-admin/pkg/auth"
	"github.com/nostack-cn/svr-admin/service"
)

// Handlers 聚合所有处理器
type Handlers struct {
	Auth  *handler.AuthHandler
	Admin *handler.AdminHandler
	RBAC  *handler.RBACHandler
	User  *handler.UserHandler
	Order *handler.OrderHandler
	Log   *handler.LogHandler
}

// Setup 注册所有路由
func Setup(r *gin.Engine, h *Handlers, jwtMgr *auth.JWTManager, auditSvc *service.AuditService) {
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())

	// 健康检查
	r.GET("/health", handler.HealthCheck)

	// 简写
	audit := func(action string) gin.HandlerFunc { return middleware.Audit(auditSvc, action) }
	perm := middleware.RequirePermission

	admin := r.Group("/api/v1/admin")
	{
		// --- 认证（免登录） ---
		authGroup := admin.Group("/auth")
		{
			authGroup.POST("/login", h.Auth.Login)
			authGroup.POST("/refresh", h.Auth.RefreshToken)
		}

		// --- 需要登录 ---
		authed := admin.Group("")
		authed.Use(middleware.JWTAuth(jwtMgr))
		{
			// 自身
			authed.POST("/auth/logout", h.Auth.Logout)
			authed.GET("/profile", h.Auth.GetProfile)
			authed.PUT("/profile/password", h.Auth.ChangePassword)

			// 管理员管理
			authed.GET("/admins", perm(auth.PermAdminRead), h.Admin.List)
			authed.GET("/admins/:id", perm(auth.PermAdminRead), h.Admin.Get)
			authed.POST("/admins", perm(auth.PermAdminWrite), audit("admin.create"), h.Admin.Create)
			authed.PUT("/admins/:id", perm(auth.PermAdminWrite), audit("admin.update"), h.Admin.Update)
			authed.POST("/admins/:id/status", perm(auth.PermAdminWrite), audit("admin.set_status"), h.Admin.SetStatus)
			authed.POST("/admins/:id/reset-password", perm(auth.PermAdminWrite), audit("admin.reset_password"), h.Admin.ResetPassword)

			// 角色与权限
			authed.GET("/permissions", perm(auth.PermPermRead), h.RBAC.ListPermissions)
			authed.GET("/roles", perm(auth.PermRoleRead), h.RBAC.ListRoles)
			authed.GET("/roles/:id", perm(auth.PermRoleRead), h.RBAC.GetRole)
			authed.POST("/roles", perm(auth.PermRoleWrite), audit("role.create"), h.RBAC.CreateRole)
			authed.PUT("/roles/:id", perm(auth.PermRoleWrite), audit("role.update"), h.RBAC.UpdateRole)
			authed.DELETE("/roles/:id", perm(auth.PermRoleWrite), audit("role.delete"), h.RBAC.DeleteRole)
			authed.PUT("/roles/:id/permissions", perm(auth.PermRoleWrite), audit("role.set_permissions"), h.RBAC.SetRolePermissions)

			// C 端用户管控（编排 svr-profile）
			authed.GET("/users", perm(auth.PermUserRead), h.User.List)
			authed.GET("/users/:id", perm(auth.PermUserRead), h.User.Get)
			authed.POST("/users/:id/ban", perm(auth.PermUserBan), audit("user.ban"), h.User.Ban)
			authed.POST("/users/:id/unban", perm(auth.PermUserBan), audit("user.unban"), h.User.Unban)
			authed.POST("/users/:id/reset-password", perm(auth.PermUserResetPassword), audit("user.reset_password"), h.User.ResetPassword)

			// 订单管控
			authed.GET("/orders", perm(auth.PermOrderRead), h.Order.List)
			authed.GET("/orders/:id", perm(auth.PermOrderRead), h.Order.Get)
			authed.POST("/orders/:id/refund", perm(auth.PermOrderRefund), audit("order.refund"), h.Order.Refund)

			// 订阅管控
			authed.POST("/subscriptions/cancel", perm(auth.PermSubscriptionCancel), audit("subscription.cancel"), h.Order.CancelSubscription)

			// 操作日志
			authed.GET("/logs", perm(auth.PermLogRead), h.Log.List)
		}
	}
}
