package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/nostack-cn/svr-admin/model"
	"github.com/nostack-cn/svr-admin/pkg/pagination"
	"github.com/nostack-cn/svr-admin/pkg/profile"
	"github.com/nostack-cn/svr-admin/pkg/response"
	"github.com/nostack-cn/svr-admin/service"
)

// OrderHandler 订单管控处理器
type OrderHandler struct {
	orderAdmin *service.OrderAdminService
}

// NewOrderHandler 创建订单管控处理器
func NewOrderHandler(s *service.OrderAdminService) *OrderHandler {
	return &OrderHandler{orderAdmin: s}
}

// List 订单列表
// GET /api/v1/admin/orders?user_id=&status=&page=&page_size=
func (h *OrderHandler) List(c *gin.Context) {
	p := pagination.Parse(c)
	var userID uint
	if v := c.Query("user_id"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			userID = uint(n)
		}
	}
	result, err := h.orderAdmin.ListOrders(c.Request.Context(), profile.ListOrdersParams{
		UserID:   userID,
		Status:   c.Query("status"),
		Page:     p.Page,
		PageSize: p.PageSize,
	})
	if err != nil {
		response.BusinessError(c, 10401, err.Error())
		return
	}
	response.OK(c, result)
}

// Get 订单详情
// GET /api/v1/admin/orders/:id
func (h *OrderHandler) Get(c *gin.Context) {
	id := parseUintParam(c, "id")
	if id == 0 {
		response.BadRequest(c, "无效的订单 ID")
		return
	}
	order, err := h.orderAdmin.GetOrder(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, order)
}

// RefundRequest 退款请求
type RefundRequest struct {
	Amount     int64  `json:"amount" binding:"required,gt=0"` // 退款金额（分）
	RefundDesc string `json:"refund_desc"`
	CancelSub  bool   `json:"cancel_sub"` // 是否同时取消用户订阅
}

// Refund 退款（支持部分退款 + 可选取消订阅）
// POST /api/v1/admin/orders/:id/refund
func (h *OrderHandler) Refund(c *gin.Context) {
	id := parseUintParam(c, "id")
	if id == 0 {
		response.BadRequest(c, "无效的订单 ID")
		return
	}
	var req RefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	refund, err := h.orderAdmin.RefundAndCancelSubscription(
		c.Request.Context(), id, req.Amount, req.RefundDesc, req.CancelSub)
	if err != nil {
		c.Set("audit_result", string(model.OperationResultFail))
		c.Set("audit_error", err.Error())
		response.BusinessError(c, 10402, err.Error())
		return
	}
	c.Set("audit_detail", "amount="+strconv.FormatInt(req.Amount, 10)+",cancel_sub="+strconv.FormatBool(req.CancelSub))
	response.OK(c, refund)
}

// CancelSubscriptionRequest 取消订阅请求
type CancelSubscriptionRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	Reason string `json:"reason"`
}

// CancelSubscription 取消用户订阅（不退款）
// POST /api/v1/admin/subscriptions/cancel
func (h *OrderHandler) CancelSubscription(c *gin.Context) {
	var req CancelSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.orderAdmin.CancelSubscription(c.Request.Context(), req.UserID, req.Reason); err != nil {
		c.Set("audit_result", string(model.OperationResultFail))
		c.Set("audit_error", err.Error())
		response.BusinessError(c, 10403, err.Error())
		return
	}
	c.Set("audit_resource", strconv.FormatUint(uint64(req.UserID), 10))
	response.OKMsg(c, "订阅已取消", nil)
}
