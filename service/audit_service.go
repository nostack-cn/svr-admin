package service

import (
	"gorm.io/gorm"

	"github.com/nostack-cn/svr-admin/model"
	"github.com/nostack-cn/svr-admin/pkg/pagination"
)

// AuditService 操作审计服务
type AuditService struct {
	db *gorm.DB
}

// NewAuditService 创建审计服务
func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{db: db}
}

// Record 写入一条操作日志（失败不影响主流程，仅记录返回错误）
func (s *AuditService) Record(log *model.AdminOperationLog) error {
	return s.db.Create(log).Error
}

// ListParams 日志查询过滤条件
type ListParams struct {
	AdminID uint
	Action  string
	Result  string
}

// List 分页查询操作日志
func (s *AuditService) List(p pagination.Params, f ListParams) ([]model.AdminOperationLog, int64, error) {
	q := s.db.Model(&model.AdminOperationLog{})
	if f.AdminID > 0 {
		q = q.Where("admin_id = ?", f.AdminID)
	}
	if f.Action != "" {
		q = q.Where("action = ?", f.Action)
	}
	if f.Result != "" {
		q = q.Where("result = ?", f.Result)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var logs []model.AdminOperationLog
	if err := q.Order("id DESC").Offset(p.Offset()).Limit(p.Limit()).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}
