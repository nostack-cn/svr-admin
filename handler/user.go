package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/nostack-cn/svr-admin/model"
	"github.com/nostack-cn/svr-admin/pkg/pagination"
	"github.com/nostack-cn/svr-admin/pkg/profile"
	"github.com/nostack-cn/svr-admin/pkg/response"
	"github.com/nostack-cn/svr-admin/service"
)

// UserHandler C 端用户管控处理器（编排 svr-profile）
type UserHandler struct {
	userAdmin *service.UserAdminService
}

// NewUserHandler 创建用户管控处理器
func NewUserHandler(s *service.UserAdminService) *UserHandler {
	return &UserHandler{userAdmin: s}
}

// List 用户列表
// GET /api/v1/admin/users
func (h *UserHandler) List(c *gin.Context) {
	p := pagination.Parse(c)
	res, err := h.userAdmin.ListUsers(c.Request.Context(), profile.ListUsersParams{
		Page:     p.Page,
		PageSize: p.PageSize,
		Keyword:  c.Query("keyword"),
		Status:   c.Query("status"),
	})
	if err != nil {
		response.BusinessError(c, 10301, err.Error())
		return
	}
	response.OKPage(c, res.List, res.Total, p.Page, p.PageSize)
}

// Get 用户详情
// GET /api/v1/admin/users/:id
func (h *UserHandler) Get(c *gin.Context) {
	id := parseUintParam(c, "id")
	if id == 0 {
		response.BadRequest(c, "无效的用户 ID")
		return
	}
	user, err := h.userAdmin.GetUser(c.Request.Context(), id)
	if err != nil {
		response.BusinessError(c, 10302, err.Error())
		return
	}
	response.OK(c, user)
}

// Ban 封禁用户
// POST /api/v1/admin/users/:id/ban
func (h *UserHandler) Ban(c *gin.Context) {
	id := parseUintParam(c, "id")
	if id == 0 {
		response.BadRequest(c, "无效的用户 ID")
		return
	}
	if err := h.userAdmin.BanUser(c.Request.Context(), id); err != nil {
		c.Set("audit_result", string(model.OperationResultFail))
		c.Set("audit_error", err.Error())
		response.BusinessError(c, 10303, err.Error())
		return
	}
	response.OKMsg(c, "已封禁", nil)
}

// Unban 解封用户
// POST /api/v1/admin/users/:id/unban
func (h *UserHandler) Unban(c *gin.Context) {
	id := parseUintParam(c, "id")
	if id == 0 {
		response.BadRequest(c, "无效的用户 ID")
		return
	}
	if err := h.userAdmin.UnbanUser(c.Request.Context(), id); err != nil {
		c.Set("audit_result", string(model.OperationResultFail))
		c.Set("audit_error", err.Error())
		response.BusinessError(c, 10304, err.Error())
		return
	}
	response.OKMsg(c, "已解封", nil)
}

// ResetPassword 重置用户密码
// POST /api/v1/admin/users/:id/reset-password
func (h *UserHandler) ResetPassword(c *gin.Context) {
	id := parseUintParam(c, "id")
	if id == 0 {
		response.BadRequest(c, "无效的用户 ID")
		return
	}
	res, err := h.userAdmin.ResetUserPassword(c.Request.Context(), id)
	if err != nil {
		c.Set("audit_result", string(model.OperationResultFail))
		c.Set("audit_error", err.Error())
		response.BusinessError(c, 10305, err.Error())
		return
	}
	response.OK(c, res)
}
