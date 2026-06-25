package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/nostack-cn/svr-admin/model"
	"github.com/nostack-cn/svr-admin/pkg/pagination"
	"github.com/nostack-cn/svr-admin/pkg/response"
	"github.com/nostack-cn/svr-admin/service"
)

// AdminHandler 管理员管理处理器
type AdminHandler struct {
	adminService *service.AdminService
}

// NewAdminHandler 创建管理员管理处理器
func NewAdminHandler(s *service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: s}
}

// List 管理员列表
// GET /api/v1/admin/admins
func (h *AdminHandler) List(c *gin.Context) {
	p := pagination.Parse(c)
	list, total, err := h.adminService.List(p, c.Query("keyword"))
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OKPage(c, list, total, p.Page, p.PageSize)
}

// Get 管理员详情
// GET /api/v1/admin/admins/:id
func (h *AdminHandler) Get(c *gin.Context) {
	id := parseUintParam(c, "id")
	if id == 0 {
		response.BadRequest(c, "无效的管理员 ID")
		return
	}
	admin, err := h.adminService.GetByID(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, admin)
}

// Create 创建管理员
// POST /api/v1/admin/admins
func (h *AdminHandler) Create(c *gin.Context) {
	var req service.CreateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	admin, err := h.adminService.Create(&req)
	if err != nil {
		c.Set("audit_result", string(model.OperationResultFail))
		c.Set("audit_error", err.Error())
		response.BusinessError(c, 10101, err.Error())
		return
	}
	response.OK(c, admin)
}

// Update 更新管理员
// PUT /api/v1/admin/admins/:id
func (h *AdminHandler) Update(c *gin.Context) {
	id := parseUintParam(c, "id")
	if id == 0 {
		response.BadRequest(c, "无效的管理员 ID")
		return
	}
	var req service.UpdateAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	admin, err := h.adminService.Update(id, &req)
	if err != nil {
		c.Set("audit_result", string(model.OperationResultFail))
		c.Set("audit_error", err.Error())
		response.BusinessError(c, 10102, err.Error())
		return
	}
	response.OK(c, admin)
}

// SetStatusRequest 设置状态请求
type SetStatusRequest struct {
	Status model.AdminStatus `json:"status" binding:"required"`
}

// SetStatus 启用/禁用管理员
// POST /api/v1/admin/admins/:id/status
func (h *AdminHandler) SetStatus(c *gin.Context) {
	id := parseUintParam(c, "id")
	if id == 0 {
		response.BadRequest(c, "无效的管理员 ID")
		return
	}
	var req SetStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.adminService.SetStatus(id, req.Status); err != nil {
		c.Set("audit_result", string(model.OperationResultFail))
		c.Set("audit_error", err.Error())
		response.BusinessError(c, 10103, err.Error())
		return
	}
	response.OKMsg(c, "操作成功", nil)
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=8,max=64"`
}

// ResetPassword 重置管理员密码
// POST /api/v1/admin/admins/:id/reset-password
func (h *AdminHandler) ResetPassword(c *gin.Context) {
	id := parseUintParam(c, "id")
	if id == 0 {
		response.BadRequest(c, "无效的管理员 ID")
		return
	}
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.adminService.ResetPassword(id, req.NewPassword); err != nil {
		c.Set("audit_result", string(model.OperationResultFail))
		c.Set("audit_error", err.Error())
		response.BusinessError(c, 10104, err.Error())
		return
	}
	response.OKMsg(c, "密码重置成功", nil)
}
