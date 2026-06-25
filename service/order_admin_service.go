package service

import (
	"context"
	"errors"

	"github.com/nostack-cn/svr-admin/pkg/consoleclient"
	"github.com/nostack-cn/svr-admin/pkg/profile"
)

// OrderAdminService 订单管控服务（编排 svr-profile + svr-console）
type OrderAdminService struct {
	profileClient *profile.Client
	consoleClient *consoleclient.Client
}

// NewOrderAdminService 创建订单管控服务
func NewOrderAdminService(pc *profile.Client, cc *consoleclient.Client) *OrderAdminService {
	return &OrderAdminService{profileClient: pc, consoleClient: cc}
}

var (
	errProfileNA = errors.New("svr-profile 下游未配置，订单管控不可用")
	errConsoleNA = errors.New("svr-console 下游未配置，订阅取消不可用")
)

// ListOrders 分页查询订单
func (s *OrderAdminService) ListOrders(ctx context.Context, p profile.ListOrdersParams) (*profile.OrderListResult, error) {
	if s.profileClient == nil {
		return nil, errProfileNA
	}
	return s.profileClient.ListOrders(ctx, p)
}

// GetOrder 查询订单详情
func (s *OrderAdminService) GetOrder(ctx context.Context, orderID uint) (*profile.Order, error) {
	if s.profileClient == nil {
		return nil, errProfileNA
	}
	return s.profileClient.GetOrder(ctx, orderID)
}

// RefundOrder 退款（支持部分退款）
func (s *OrderAdminService) RefundOrder(ctx context.Context, orderID uint, amount int64, desc string) (*profile.Refund, error) {
	if s.profileClient == nil {
		return nil, errProfileNA
	}
	return s.profileClient.RefundOrder(ctx, orderID, &profile.RefundRequest{
		Amount:     amount,
		RefundDesc: desc,
	})
}

// RefundAndCancelSubscription 全额退款 + 取消关联用户订阅
func (s *OrderAdminService) RefundAndCancelSubscription(ctx context.Context, orderID uint, amount int64, desc string, cancelSub bool) (*profile.Refund, error) {
	refund, err := s.RefundOrder(ctx, orderID, amount, desc)
	if err != nil {
		return nil, err
	}

	// 退款成功后，如需取消订阅
	if cancelSub && s.consoleClient != nil {
		// 先查订单获取 user_id
		order, _ := s.profileClient.GetOrder(ctx, orderID)
		if order != nil {
			_ = s.consoleClient.CancelSubscription(ctx, order.UserID, "管理员退款取消订阅: "+desc)
		}
	}
	return refund, nil
}

// CancelSubscription 仅取消订阅（不退款）
func (s *OrderAdminService) CancelSubscription(ctx context.Context, userID uint, reason string) error {
	if s.consoleClient == nil {
		return errConsoleNA
	}
	return s.consoleClient.CancelSubscription(ctx, userID, reason)
}
