package model

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel 通用基础模型，所有业务模型可嵌入
type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// AdminStatus 管理员状态枚举
type AdminStatus string

const (
	AdminStatusActive   AdminStatus = "active"   // 正常
	AdminStatusDisabled AdminStatus = "disabled" // 禁用
)

// Admin 管理员账号
type Admin struct {
	BaseModel
	Username    string      `gorm:"type:varchar(64);uniqueIndex;size:64;not null" json:"username"`
	Password    string      `gorm:"type:varchar(255);not null" json:"-"`
	Name        string      `gorm:"type:varchar(64);size:64" json:"name"`
	Email       string      `gorm:"type:varchar(255);uniqueIndex;size:255" json:"email"`
	Status      AdminStatus `gorm:"type:varchar(20);default:'active';index" json:"status"`
	RoleID      uint        `gorm:"index;not null" json:"role_id"`
	Role        *Role       `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	LastLoginAt *time.Time  `json:"last_login_at,omitempty"`
	LastLoginIP string      `gorm:"type:varchar(64);size:64" json:"last_login_ip,omitempty"`
}

// TableName 指定表名
func (Admin) TableName() string {
	return "admins"
}

// Role 角色
type Role struct {
	BaseModel
	Code        string `gorm:"type:varchar(64);uniqueIndex;size:64;not null" json:"code"`
	Name        string `gorm:"type:varchar(64);size:64;not null" json:"name"`
	Description string `gorm:"type:varchar(255);size:255" json:"description"`
	IsSystem    bool   `gorm:"default:false" json:"is_system"` // 系统内置角色，不可删除
	// Permissions 关联权限点（多对多，通过 role_permissions 关联表）
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

// TableName 指定表名
func (Role) TableName() string {
	return "roles"
}

// Permission 权限点
type Permission struct {
	BaseModel
	Code  string `gorm:"type:varchar(128);uniqueIndex;size:128;not null" json:"code"` // service:resource:action
	Name  string `gorm:"type:varchar(64);size:64;not null" json:"name"`
	Group string `gorm:"type:varchar(64);size:64;index" json:"group"` // 分组展示
}

// TableName 指定表名
func (Permission) TableName() string {
	return "permissions"
}

// OperationResult 操作结果枚举
type OperationResult string

const (
	OperationResultSuccess OperationResult = "success"
	OperationResultFail    OperationResult = "fail"
)

// AdminOperationLog 操作审计日志（所有写操作必记）
type AdminOperationLog struct {
	BaseModel
	AdminID   uint            `gorm:"index" json:"admin_id"`
	AdminName string          `gorm:"type:varchar(64);size:64" json:"admin_name"` // 冗余便于查询
	Action    string          `gorm:"type:varchar(128);size:128;index" json:"action"`
	Resource  string          `gorm:"type:varchar(128);size:128" json:"resource"` // 目标资源，如 user:1001
	Detail    string          `gorm:"type:text" json:"detail,omitempty"`          // JSON 摘要
	IP        string          `gorm:"type:varchar(64);size:64" json:"ip,omitempty"`
	Result    OperationResult `gorm:"type:varchar(20);default:'success';index" json:"result"`
	ErrorMsg  string          `gorm:"type:text" json:"error_msg,omitempty"`
}

// TableName 指定表名
func (AdminOperationLog) TableName() string {
	return "admin_operation_logs"
}
