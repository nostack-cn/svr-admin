package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/nostack-cn/svr-admin/middleware"
	"github.com/nostack-cn/svr-admin/pkg/response"
	"github.com/nostack-cn/svr-admin/service"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	adminService *service.AdminService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(s *service.AdminService) *AuthHandler {
	return &AuthHandler{adminService: s}
}

// Login 管理员登录
// POST /api/v1/admin/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := h.adminService.Login(&req, c.ClientIP())
	if err != nil {
		response.BusinessError(c, 10001, err.Error())
		return
	}
	response.OK(c, result)
}

// RefreshToken 刷新 Token
// POST /api/v1/admin/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req service.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	result, err := h.adminService.RefreshToken(&req)
	if err != nil {
		response.BusinessError(c, 10002, err.Error())
		return
	}
	response.OK(c, result)
}

// Logout 登出（无状态，客户端丢弃 Token 即可；Access Token 短 TTL 自动失效）
// POST /api/v1/admin/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	response.OKMsg(c, "登出成功", nil)
}

// GetProfile 获取当前管理员信息
// GET /api/v1/admin/profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	admin, err := h.adminService.GetByID(middleware.GetAdminID(c))
	if err != nil {
		response.BusinessError(c, 10003, err.Error())
		return
	}
	response.OK(c, admin)
}

// ChangePassword 修改自身密码
// PUT /api/v1/admin/profile/password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req service.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.adminService.ChangePassword(middleware.GetAdminID(c), &req); err != nil {
		response.BusinessError(c, 10004, err.Error())
		return
	}
	response.OKMsg(c, "密码修改成功", nil)
}
