package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/nostack-cn/svr-admin/model"
	"github.com/nostack-cn/svr-admin/pkg/response"
	"github.com/nostack-cn/svr-admin/service"
)

// RBACHandler 角色权限处理器
type RBACHandler struct {
	rbac *service.RBACService
}

// NewRBACHandler 创建角色权限处理器
func NewRBACHandler(s *service.RBACService) *RBACHandler {
	return &RBACHandler{rbac: s}
}

// ListPermissions 权限点列表
// GET /api/v1/admin/permissions
func (h *RBACHandler) ListPermissions(c *gin.Context) {
	perms, err := h.rbac.ListPermissions()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, perms)
}

// ListRoles 角色列表
// GET /api/v1/admin/roles
func (h *RBACHandler) ListRoles(c *gin.Context) {
	roles, err := h.rbac.ListRoles()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OK(c, roles)
}

// GetRole 角色详情
// GET /api/v1/admin/roles/:id
func (h *RBACHandler) GetRole(c *gin.Context) {
	id := parseUintParam(c, "id")
	if id == 0 {
		response.BadRequest(c, "无效的角色 ID")
		return
	}
	role, err := h.rbac.GetRole(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, role)
}

// CreateRole 创建角色
// POST /api/v1/admin/roles
func (h *RBACHandler) CreateRole(c *gin.Context) {
	var req service.CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	role, err := h.rbac.CreateRole(&req)
	if err != nil {
		c.Set("audit_result", string(model.OperationResultFail))
		c.Set("audit_error", err.Error())
		response.BusinessError(c, 10201, err.Error())
		return
	}
	response.OK(c, role)
}

// UpdateRole 更新角色
// PUT /api/v1/admin/roles/:id
func (h *RBACHandler) UpdateRole(c *gin.Context) {
	id := parseUintParam(c, "id")
	if id == 0 {
		response.BadRequest(c, "无效的角色 ID")
		return
	}
	var req service.UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	role, err := h.rbac.UpdateRole(id, &req)
	if err != nil {
		c.Set("audit_result", string(model.OperationResultFail))
		c.Set("audit_error", err.Error())
		response.BusinessError(c, 10202, err.Error())
		return
	}
	response.OK(c, role)
}

// DeleteRole 删除角色
// DELETE /api/v1/admin/roles/:id
func (h *RBACHandler) DeleteRole(c *gin.Context) {
	id := parseUintParam(c, "id")
	if id == 0 {
		response.BadRequest(c, "无效的角色 ID")
		return
	}
	if err := h.rbac.DeleteRole(id); err != nil {
		c.Set("audit_result", string(model.OperationResultFail))
		c.Set("audit_error", err.Error())
		response.BusinessError(c, 10203, err.Error())
		return
	}
	response.OKMsg(c, "删除成功", nil)
}

// SetRolePermissionsRequest 设置角色权限请求
type SetRolePermissionsRequest struct {
	PermissionCodes []string `json:"permission_codes"`
}

// SetRolePermissions 全量设置角色权限
// PUT /api/v1/admin/roles/:id/permissions
func (h *RBACHandler) SetRolePermissions(c *gin.Context) {
	id := parseUintParam(c, "id")
	if id == 0 {
		response.BadRequest(c, "无效的角色 ID")
		return
	}
	var req SetRolePermissionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	role, err := h.rbac.SetRolePermissions(id, req.PermissionCodes)
	if err != nil {
		c.Set("audit_result", string(model.OperationResultFail))
		c.Set("audit_error", err.Error())
		response.BusinessError(c, 10204, err.Error())
		return
	}
	response.OK(c, role)
}
