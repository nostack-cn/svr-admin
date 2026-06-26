package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nostack-cn/svr-admin/model"
	"github.com/nostack-cn/svr-admin/pkg/auth"
	"github.com/nostack-cn/svr-admin/pkg/response"
	"github.com/nostack-cn/svr-admin/service"
)

// JWTAuth 管理员 JWT 认证中间件（无状态，仅校验签名与有效期）
func JWTAuth(jwtMgr *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "缺少认证信息")
			c.Abort()
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "认证格式错误")
			c.Abort()
			return
		}
		claims, err := jwtMgr.ParseAccessToken(parts[1])
		if err != nil {
			response.Unauthorized(c, "Token 无效或已过期")
			c.Abort()
			return
		}

		c.Set("admin_id", claims.AdminID)
		c.Set("admin_username", claims.Username)
		c.Set("admin_role", claims.Role)
		c.Set("admin_permissions", claims.Permissions)
		c.Next()
	}
}

// RequirePermission 权限点校验中间件（super_admin 跳过）
func RequirePermission(perm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if GetAdminRole(c) == auth.RoleSuperAdmin {
			c.Next()
			return
		}
		for _, p := range GetAdminPermissions(c) {
			if p == perm {
				c.Next()
				return
			}
		}
		response.Forbidden(c, "无权访问该资源")
		c.Abort()
	}
}

// RequireSuperAdmin 仅超级管理员可访问
func RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if GetAdminRole(c) == auth.RoleSuperAdmin {
			c.Next()
			return
		}
		response.Forbidden(c, "需要超级管理员权限")
		c.Abort()
	}
}

// Audit 操作审计中间件：在 handler 执行后记录一条操作日志。
// handler 可通过 c.Set 注入细节：audit_resource / audit_detail / audit_error。
func Audit(auditSvc *service.AuditService, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		result := model.OperationResultSuccess
		if c.Writer.Status() >= http.StatusBadRequest {
			result = model.OperationResultFail
		}
		if v, ok := c.Get("audit_result"); ok {
			if s, ok := v.(string); ok && s == string(model.OperationResultFail) {
				result = model.OperationResultFail
			}
		}

		resource := c.Param("id")
		if v, ok := c.Get("audit_resource"); ok {
			if s, ok := v.(string); ok {
				resource = s
			}
		}
		detail, _ := c.Get("audit_detail")
		detailStr, _ := detail.(string)
		errVal, _ := c.Get("audit_error")
		errStr, _ := errVal.(string)

		_ = auditSvc.Record(&model.AdminOperationLog{
			AdminID:   GetAdminID(c),
			AdminName: GetAdminUsername(c),
			Action:    action,
			Resource:  resource,
			Detail:    detailStr,
			IP:        c.ClientIP(),
			Result:    result,
			ErrorMsg:  errStr,
		})
	}
}

// GetAdminID 从上下文获取管理员 ID
func GetAdminID(c *gin.Context) uint {
	if v, ok := c.Get("admin_id"); ok {
		if id, ok := v.(uint); ok {
			return id
		}
	}
	return 0
}

// GetAdminUsername 从上下文获取管理员用户名
func GetAdminUsername(c *gin.Context) string {
	if v, ok := c.Get("admin_username"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetAdminRole 从上下文获取管理员角色码
func GetAdminRole(c *gin.Context) string {
	if v, ok := c.Get("admin_role"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// GetAdminPermissions 从上下文获取管理员权限码列表
func GetAdminPermissions(c *gin.Context) []string {
	if v, ok := c.Get("admin_permissions"); ok {
		if ps, ok := v.([]string); ok {
			return ps
		}
	}
	return nil
}

// InternalAuth 内部服务鉴权中间件（X-Internal-Key）
func InternalAuth(internalKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-Internal-Key")
		if key == "" || internalKey == "" || key != internalKey {
			response.Unauthorized(c, "内部服务鉴权失败")
			c.Abort()
			return
		}
		c.Next()
	}
}

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,PATCH,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization,X-Requested-With")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// Logger 请求日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method

		_ = latency
		_ = path
		_ = statusCode
		_ = method
	}
}
