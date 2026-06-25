package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/nostack-cn/svr-admin/pkg/pagination"
	"github.com/nostack-cn/svr-admin/pkg/response"
	"github.com/nostack-cn/svr-admin/service"
)

// LogHandler 操作日志处理器
type LogHandler struct {
	audit *service.AuditService
}

// NewLogHandler 创建操作日志处理器
func NewLogHandler(s *service.AuditService) *LogHandler {
	return &LogHandler{audit: s}
}

// List 操作日志列表
// GET /api/v1/admin/logs
func (h *LogHandler) List(c *gin.Context) {
	p := pagination.Parse(c)
	var adminID uint
	if v := c.Query("admin_id"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			adminID = uint(n)
		}
	}
	logs, total, err := h.audit.List(p, service.ListParams{
		AdminID: adminID,
		Action:  c.Query("action"),
		Result:  c.Query("result"),
	})
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.OKPage(c, logs, total, p.Page, p.PageSize)
}
