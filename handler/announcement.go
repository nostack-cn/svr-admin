package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/nostack-cn/svr-admin/middleware"
	"github.com/nostack-cn/svr-admin/model"
	"github.com/nostack-cn/svr-admin/pkg/pagination"
	"github.com/nostack-cn/svr-admin/pkg/response"
	"github.com/nostack-cn/svr-admin/service"
)

// AnnouncementHandler 公告处理器
type AnnouncementHandler struct {
	svc *service.AnnouncementService
}

// NewAnnouncementHandler 创建公告处理器
func NewAnnouncementHandler(svc *service.AnnouncementService) *AnnouncementHandler {
	return &AnnouncementHandler{svc: svc}
}

// ======================== 管理接口（需 JWT + 权限） ========================

// List GET /api/v1/admin/announcements
func (h *AnnouncementHandler) List(c *gin.Context) {
	p := pagination.Parse(c)
	list, total, err := h.svc.ListAll(
		p.Page, p.PageSize,
		c.Query("type"),
		c.Query("status"),
		c.Query("keyword"),
	)
	if err != nil {
		response.InternalError(c, "查询公告列表失败")
		return
	}
	response.OKPage(c, list, total, p.Page, p.PageSize)
}

// Get GET /api/v1/admin/announcements/:id
func (h *AnnouncementHandler) Get(c *gin.Context) {
	id := parseUint(c, "id")
	if id == 0 {
		response.BadRequest(c, "无效的公告 ID")
		return
	}
	a, err := h.svc.GetByID(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, a)
}

// Create POST /api/v1/admin/announcements
func (h *AnnouncementHandler) Create(c *gin.Context) {
	var req service.CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	req.AuthorID = middleware.GetAdminID(c)
	req.AuthorName = middleware.GetAdminUsername(c)

	a, err := h.svc.Create(&req)
	if err != nil {
		response.BusinessError(c, 60001, err.Error())
		return
	}
	c.Set("audit_resource", "announcement:"+formatUint(a.ID))
	c.Set("audit_detail", a.Title)
	response.OK(c, a)
}

// Update PUT /api/v1/admin/announcements/:id
func (h *AnnouncementHandler) Update(c *gin.Context) {
	id := parseUint(c, "id")
	if id == 0 {
		response.BadRequest(c, "无效的公告 ID")
		return
	}
	var req service.UpdateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	a, err := h.svc.Update(id, &req)
	if err != nil {
		response.BusinessError(c, 60002, err.Error())
		return
	}
	c.Set("audit_resource", "announcement:"+formatUint(id))
	c.Set("audit_detail", a.Title)
	response.OK(c, a)
}

// Delete DELETE /api/v1/admin/announcements/:id
func (h *AnnouncementHandler) Delete(c *gin.Context) {
	id := parseUint(c, "id")
	if id == 0 {
		response.BadRequest(c, "无效的公告 ID")
		return
	}
	if err := h.svc.Delete(id); err != nil {
		response.BusinessError(c, 60003, err.Error())
		return
	}
	c.Set("audit_resource", "announcement:"+formatUint(id))
	response.OKMsg(c, "已删除", nil)
}

// ======================== 内部接口（X-Internal-Key 鉴权） ========================

// InternalListSite GET /internal/announcements/site
// 返回 web-site 有效已发布公告
func (h *AnnouncementHandler) InternalListSite(c *gin.Context) {
	list, err := h.svc.ListActiveByType("site")
	if err != nil {
		response.InternalError(c, "查询公告失败")
		return
	}
	if list == nil {
		list = []model.Announcement{}
	}
	response.OK(c, list)
}

// InternalListConsole GET /internal/announcements/console
// 返回 web-console 有效已发布公告
func (h *AnnouncementHandler) InternalListConsole(c *gin.Context) {
	list, err := h.svc.ListActiveByType("console")
	if err != nil {
		response.InternalError(c, "查询公告失败")
		return
	}
	if list == nil {
		list = []model.Announcement{}
	}
	response.OK(c, list)
}
