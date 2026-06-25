package service

import (
	"context"
	"errors"

	"github.com/nostack-cn/svr-admin/pkg/profile"
)

// UserAdminService 用户管控服务（编排下游 svr-profile internal API）
type UserAdminService struct {
	client *profile.Client
}

// NewUserAdminService 创建用户管控服务。client 为 nil 时（未配置 svr-profile）相关操作返回错误。
func NewUserAdminService(client *profile.Client) *UserAdminService {
	return &UserAdminService{client: client}
}

var errProfileUnavailable = errors.New("svr-profile 下游未配置，用户管控不可用")

// ListUsers 分页查询用户
func (s *UserAdminService) ListUsers(ctx context.Context, p profile.ListUsersParams) (*profile.ListUsersResult, error) {
	if s.client == nil {
		return nil, errProfileUnavailable
	}
	return s.client.ListUsers(ctx, p)
}

// GetUser 查询单个用户
func (s *UserAdminService) GetUser(ctx context.Context, id uint) (*profile.User, error) {
	if s.client == nil {
		return nil, errProfileUnavailable
	}
	return s.client.GetUser(ctx, id)
}

// BanUser 封禁用户
func (s *UserAdminService) BanUser(ctx context.Context, id uint) error {
	if s.client == nil {
		return errProfileUnavailable
	}
	return s.client.SetUserStatus(ctx, id, "banned")
}

// UnbanUser 解封用户
func (s *UserAdminService) UnbanUser(ctx context.Context, id uint) error {
	if s.client == nil {
		return errProfileUnavailable
	}
	return s.client.SetUserStatus(ctx, id, "active")
}

// ResetUserPassword 重置用户密码，返回新密码（由 svr-profile 生成）
func (s *UserAdminService) ResetUserPassword(ctx context.Context, id uint) (*profile.ResetPasswordResult, error) {
	if s.client == nil {
		return nil, errProfileUnavailable
	}
	return s.client.ResetUserPassword(ctx, id)
}
