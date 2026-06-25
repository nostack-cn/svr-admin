package service

import (
	"errors"

	"gorm.io/gorm"

	"github.com/nostack-cn/svr-admin/model"
	"github.com/nostack-cn/svr-admin/pkg/auth"
)

// RBACService 角色与权限服务
type RBACService struct {
	db *gorm.DB
}

// NewRBACService 创建 RBAC 服务
func NewRBACService(db *gorm.DB) *RBACService {
	return &RBACService{db: db}
}

// ListPermissions 返回全部权限点
func (s *RBACService) ListPermissions() ([]model.Permission, error) {
	var perms []model.Permission
	if err := s.db.Order("`group`, id").Find(&perms).Error; err != nil {
		return nil, err
	}
	return perms, nil
}

// ListRoles 返回全部角色（含权限）
func (s *RBACService) ListRoles() ([]model.Role, error) {
	var roles []model.Role
	if err := s.db.Preload("Permissions").Order("id").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

// GetRole 获取角色（含权限）
func (s *RBACService) GetRole(id uint) (*model.Role, error) {
	var role model.Role
	if err := s.db.Preload("Permissions").First(&role, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("角色不存在")
		}
		return nil, err
	}
	return &role, nil
}

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// CreateRole 创建角色
func (s *RBACService) CreateRole(req *CreateRoleRequest) (*model.Role, error) {
	var count int64
	s.db.Model(&model.Role{}).Where("code = ?", req.Code).Count(&count)
	if count > 0 {
		return nil, errors.New("角色码已存在")
	}
	role := model.Role{Code: req.Code, Name: req.Name, Description: req.Description}
	if err := s.db.Create(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateRole 更新角色基础信息
func (s *RBACService) UpdateRole(id uint, req *UpdateRoleRequest) (*model.Role, error) {
	role, err := s.GetRole(id)
	if err != nil {
		return nil, err
	}
	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if len(updates) > 0 {
		if err := s.db.Model(&model.Role{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	return s.GetRole(role.ID)
}

// DeleteRole 删除角色（系统内置/已被管理员引用时禁止）
func (s *RBACService) DeleteRole(id uint) error {
	role, err := s.GetRole(id)
	if err != nil {
		return err
	}
	if role.IsSystem {
		return errors.New("系统内置角色不可删除")
	}
	var inUse int64
	s.db.Model(&model.Admin{}).Where("role_id = ?", id).Count(&inUse)
	if inUse > 0 {
		return errors.New("该角色仍有管理员在使用，不可删除")
	}
	return s.db.Delete(&model.Role{}, id).Error
}

// SetRolePermissions 全量替换角色权限（按权限码）
func (s *RBACService) SetRolePermissions(id uint, codes []string) (*model.Role, error) {
	role, err := s.GetRole(id)
	if err != nil {
		return nil, err
	}

	var perms []model.Permission
	if len(codes) > 0 {
		if err := s.db.Where("code IN ?", codes).Find(&perms).Error; err != nil {
			return nil, err
		}
		if len(perms) != len(codes) {
			return nil, errors.New("存在无效的权限码")
		}
	}

	if err := s.db.Model(role).Association("Permissions").Replace(perms); err != nil {
		return nil, err
	}
	return s.GetRole(id)
}

// GetRolePermissionCodes 返回某角色的权限码列表（登录时写入 JWT）
func (s *RBACService) GetRolePermissionCodes(roleID uint) ([]string, error) {
	role, err := s.GetRole(roleID)
	if err != nil {
		return nil, err
	}
	codes := make([]string, 0, len(role.Permissions))
	for _, p := range role.Permissions {
		codes = append(codes, p.Code)
	}
	return codes, nil
}

// SeedPermissions 同步全量权限点定义到数据库（幂等：仅插入缺失项）
func (s *RBACService) SeedPermissions() error {
	for _, def := range auth.AllPermissions() {
		var count int64
		s.db.Model(&model.Permission{}).Where("code = ?", def.Code).Count(&count)
		if count == 0 {
			if err := s.db.Create(&model.Permission{Code: def.Code, Name: def.Name, Group: def.Group}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// EnsureSuperAdminRole 确保超级管理员角色存在并拥有全部权限，返回其 ID
func (s *RBACService) EnsureSuperAdminRole() (uint, error) {
	var role model.Role
	err := s.db.Where("code = ?", auth.RoleSuperAdmin).First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		role = model.Role{Code: auth.RoleSuperAdmin, Name: "超级管理员", Description: "拥有全部权限", IsSystem: true}
		if err := s.db.Create(&role).Error; err != nil {
			return 0, err
		}
	} else if err != nil {
		return 0, err
	}

	var perms []model.Permission
	if err := s.db.Find(&perms).Error; err != nil {
		return 0, err
	}
	if err := s.db.Model(&role).Association("Permissions").Replace(perms); err != nil {
		return 0, err
	}
	return role.ID, nil
}
