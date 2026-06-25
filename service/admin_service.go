package service

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/nostack-cn/svr-admin/model"
	"github.com/nostack-cn/svr-admin/pkg/auth"
	"github.com/nostack-cn/svr-admin/pkg/pagination"
)

// AdminService 管理员服务
type AdminService struct {
	db     *gorm.DB
	jwtMgr *auth.JWTManager
	rbac   *RBACService
}

// NewAdminService 创建管理员服务
func NewAdminService(db *gorm.DB, jwtMgr *auth.JWTManager, rbac *RBACService) *AdminService {
	return &AdminService{db: db, jwtMgr: jwtMgr, rbac: rbac}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int64       `json:"expires_in"`
	TokenType    string      `json:"token_type"`
	Admin        model.Admin `json:"admin"`
}

// issueToken 为管理员签发 Token 对（加载角色码与权限快照）
func (s *AdminService) issueToken(admin *model.Admin) (*auth.TokenPair, error) {
	var role model.Role
	if err := s.db.First(&role, admin.RoleID).Error; err != nil {
		return nil, errors.New("管理员角色不存在")
	}
	perms, err := s.rbac.GetRolePermissionCodes(admin.RoleID)
	if err != nil {
		return nil, err
	}
	return s.jwtMgr.GenerateTokenPair(admin.ID, admin.Username, role.Code, perms)
}

// Login 管理员登录
func (s *AdminService) Login(req *LoginRequest, ip string) (*LoginResponse, error) {
	var admin model.Admin
	if err := s.db.Where("username = ?", req.Username).First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("用户名或密码错误")
	}
	if admin.Status != model.AdminStatusActive {
		return nil, errors.New("账号已被禁用")
	}

	pair, err := s.issueToken(&admin)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	s.db.Model(&admin).Updates(map[string]interface{}{"last_login_at": now, "last_login_ip": ip})
	admin.LastLoginAt = &now
	admin.LastLoginIP = ip

	return &LoginResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
		TokenType:    pair.TokenType,
		Admin:        admin,
	}, nil
}

// RefreshTokenRequest Token 刷新请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshToken 刷新 Access Token（重新读取最新角色/权限/状态）
func (s *AdminService) RefreshToken(req *RefreshTokenRequest) (*LoginResponse, error) {
	claims, err := s.jwtMgr.ParseRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, errors.New("Refresh Token 无效或已过期")
	}

	admin, err := s.GetByID(claims.AdminID)
	if err != nil {
		return nil, errors.New("管理员不存在")
	}
	if admin.Status != model.AdminStatusActive {
		return nil, errors.New("账号已被禁用")
	}

	pair, err := s.issueToken(admin)
	if err != nil {
		return nil, err
	}
	return &LoginResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
		TokenType:    pair.TokenType,
		Admin:        *admin,
	}, nil
}

// GetByID 根据 ID 获取管理员（含角色）
func (s *AdminService) GetByID(id uint) (*model.Admin, error) {
	var admin model.Admin
	if err := s.db.Preload("Role").First(&admin, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("管理员不存在")
		}
		return nil, err
	}
	return &admin, nil
}

// List 分页查询管理员
func (s *AdminService) List(p pagination.Params, keyword string) ([]model.Admin, int64, error) {
	q := s.db.Model(&model.Admin{}).Preload("Role")
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("username LIKE ? OR name LIKE ? OR email LIKE ?", like, like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var admins []model.Admin
	if err := q.Order("id DESC").Offset(p.Offset()).Limit(p.Limit()).Find(&admins).Error; err != nil {
		return nil, 0, err
	}
	return admins, total, nil
}

// CreateAdminRequest 创建管理员请求
type CreateAdminRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=8,max=64"`
	Name     string `json:"name"`
	Email    string `json:"email" binding:"required,email"`
	RoleID   uint   `json:"role_id" binding:"required"`
}

// Create 创建管理员
func (s *AdminService) Create(req *CreateAdminRequest) (*model.Admin, error) {
	var count int64
	s.db.Model(&model.Admin{}).Where("username = ?", req.Username).Count(&count)
	if count > 0 {
		return nil, errors.New("用户名已存在")
	}
	s.db.Model(&model.Admin{}).Where("email = ?", req.Email).Count(&count)
	if count > 0 {
		return nil, errors.New("邮箱已被使用")
	}
	if err := s.db.First(&model.Role{}, req.RoleID).Error; err != nil {
		return nil, errors.New("指定角色不存在")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}
	admin := model.Admin{
		Username: req.Username,
		Password: string(hashed),
		Name:     req.Name,
		Email:    req.Email,
		Status:   model.AdminStatusActive,
		RoleID:   req.RoleID,
	}
	if err := s.db.Create(&admin).Error; err != nil {
		return nil, err
	}
	return s.GetByID(admin.ID)
}

// UpdateAdminRequest 更新管理员请求
type UpdateAdminRequest struct {
	Name   string `json:"name"`
	Email  string `json:"email"`
	RoleID uint   `json:"role_id"`
}

// Update 更新管理员基础信息与角色
func (s *AdminService) Update(id uint, req *UpdateAdminRequest) (*model.Admin, error) {
	if _, err := s.GetByID(id); err != nil {
		return nil, err
	}
	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Email != "" {
		var count int64
		s.db.Model(&model.Admin{}).Where("email = ? AND id != ?", req.Email, id).Count(&count)
		if count > 0 {
			return nil, errors.New("邮箱已被其他管理员使用")
		}
		updates["email"] = req.Email
	}
	if req.RoleID > 0 {
		if err := s.db.First(&model.Role{}, req.RoleID).Error; err != nil {
			return nil, errors.New("指定角色不存在")
		}
		updates["role_id"] = req.RoleID
	}
	if len(updates) > 0 {
		if err := s.db.Model(&model.Admin{}).Where("id = ?", id).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	return s.GetByID(id)
}

// SetStatus 启用/禁用管理员
func (s *AdminService) SetStatus(id uint, status model.AdminStatus) error {
	if status != model.AdminStatusActive && status != model.AdminStatusDisabled {
		return errors.New("无效的状态")
	}
	if _, err := s.GetByID(id); err != nil {
		return err
	}
	return s.db.Model(&model.Admin{}).Where("id = ?", id).Update("status", status).Error
}

// ResetPassword 重置某管理员密码
func (s *AdminService) ResetPassword(id uint, newPassword string) error {
	if len(newPassword) < 8 {
		return errors.New("密码长度至少 8 位")
	}
	if _, err := s.GetByID(id); err != nil {
		return err
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}
	return s.db.Model(&model.Admin{}).Where("id = ?", id).Update("password", string(hashed)).Error
}

// ChangePasswordRequest 修改自身密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=64"`
}

// ChangePassword 管理员修改自身密码
func (s *AdminService) ChangePassword(id uint, req *ChangePasswordRequest) error {
	admin, err := s.GetByID(id)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.OldPassword)); err != nil {
		return errors.New("旧密码错误")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}
	return s.db.Model(&model.Admin{}).Where("id = ?", id).Update("password", string(hashed)).Error
}

// SeedSuperAdmin 首次启动且无任何管理员时，创建初始超级管理员
func (s *AdminService) SeedSuperAdmin(username, password, email string, roleID uint) (bool, error) {
	if password == "" {
		return false, nil // 未配置密码则跳过
	}
	var count int64
	s.db.Model(&model.Admin{}).Count(&count)
	if count > 0 {
		return false, nil
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return false, err
	}
	admin := model.Admin{
		Username: username,
		Password: string(hashed),
		Name:     "超级管理员",
		Email:    email,
		Status:   model.AdminStatusActive,
		RoleID:   roleID,
	}
	if err := s.db.Create(&admin).Error; err != nil {
		return false, err
	}
	return true, nil
}
