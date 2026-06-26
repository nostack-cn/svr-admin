package service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/nostack-cn/svr-admin/model"
)

// AnnouncementService 公告服务
type AnnouncementService struct {
	db *gorm.DB
}

// NewAnnouncementService 创建公告服务
func NewAnnouncementService(db *gorm.DB) *AnnouncementService {
	return &AnnouncementService{db: db}
}

// CreateAnnouncementRequest 创建公告请求
type CreateAnnouncementRequest struct {
	Type       string     `json:"type" binding:"required,oneof=site console"`
	Title      string     `json:"title" binding:"required"`
	Content    string     `json:"content" binding:"required"`
	Status     string     `json:"status"`
	Priority   int        `json:"priority"`
	StartsAt   *time.Time `json:"starts_at"`
	EndsAt     *time.Time `json:"ends_at"`
	AuthorID   uint       `json:"-"`
	AuthorName string     `json:"-"`
}

// UpdateAnnouncementRequest 更新公告请求
type UpdateAnnouncementRequest struct {
	Title    string     `json:"title"`
	Content  string     `json:"content"`
	Status   string     `json:"status"`
	Priority int        `json:"priority"`
	StartsAt *time.Time `json:"starts_at"`
	EndsAt   *time.Time `json:"ends_at"`
}

// Create 创建公告
func (s *AnnouncementService) Create(req *CreateAnnouncementRequest) (*model.Announcement, error) {
	status := model.AnnouncementStatusDraft
	if req.Status == string(model.AnnouncementStatusPublished) {
		status = model.AnnouncementStatusPublished
	}

	a := model.Announcement{
		Type:       model.AnnouncementType(req.Type),
		Title:      req.Title,
		Content:    req.Content,
		Status:     status,
		Priority:   req.Priority,
		StartsAt:   req.StartsAt,
		EndsAt:     req.EndsAt,
		AuthorID:   req.AuthorID,
		AuthorName: req.AuthorName,
	}

	if status == model.AnnouncementStatusPublished {
		now := time.Now()
		a.PublishedAt = &now
	}

	if err := s.db.Create(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

// Update 更新公告
func (s *AnnouncementService) Update(id uint, req *UpdateAnnouncementRequest) (*model.Announcement, error) {
	var a model.Announcement
	if err := s.db.First(&a, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("公告不存在")
		}
		return nil, err
	}

	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.Status != "" {
		newStatus := model.AnnouncementStatus(req.Status)
		if newStatus == model.AnnouncementStatusPublished && a.PublishedAt == nil {
			now := time.Now()
			updates["published_at"] = &now
		}
		updates["status"] = newStatus
	}
	if req.Priority != 0 {
		updates["priority"] = req.Priority
	}
	// StartsAt / EndsAt 允许显式置空，使用指针语义
	if req.StartsAt != nil {
		updates["starts_at"] = req.StartsAt
	}
	if req.EndsAt != nil {
		updates["ends_at"] = req.EndsAt
	}

	if len(updates) == 0 {
		return &a, nil
	}
	if err := s.db.Model(&a).Updates(updates).Error; err != nil {
		return nil, err
	}
	s.db.First(&a, id)
	return &a, nil
}

// Delete 删除公告（软删除）
func (s *AnnouncementService) Delete(id uint) error {
	result := s.db.Delete(&model.Announcement{}, id)
	if result.RowsAffected == 0 {
		return errors.New("公告不存在")
	}
	return result.Error
}

// GetByID 按 ID 获取公告
func (s *AnnouncementService) GetByID(id uint) (*model.Announcement, error) {
	var a model.Announcement
	if err := s.db.First(&a, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("公告不存在")
		}
		return nil, err
	}
	return &a, nil
}

// ListAll 分页查询公告（管理员使用）
func (s *AnnouncementService) ListAll(page, pageSize int, annoType, status, keyword string) ([]model.Announcement, int64, error) {
	var list []model.Announcement
	var total int64

	query := s.db.Model(&model.Announcement{})
	if annoType != "" {
		query = query.Where("type = ?", annoType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("title LIKE ? OR content LIKE ?", like, like)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("priority DESC, id DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// ListActiveByType 查询指定类型的有效已发布公告（内部接口使用）
// 有效 = status=published AND (starts_at IS NULL OR starts_at <= now) AND (ends_at IS NULL OR ends_at >= now)
func (s *AnnouncementService) ListActiveByType(annoType string) ([]model.Announcement, error) {
	var list []model.Announcement
	now := time.Now()

	err := s.db.Where(
		"type = ? AND status = ? AND (starts_at IS NULL OR starts_at <= ?) AND (ends_at IS NULL OR ends_at >= ?)",
		annoType, model.AnnouncementStatusPublished, now, now,
	).Order("priority DESC, published_at DESC").Find(&list).Error

	if err != nil {
		return nil, err
	}
	return list, nil
}
