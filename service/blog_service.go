package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/nostack-cn/svr-admin/model"
)

// BlogService 博客服务
type BlogService struct {
	db *gorm.DB
}

// NewBlogService 创建博客服务
func NewBlogService(db *gorm.DB) *BlogService {
	return &BlogService{db: db}
}

// CreateBlogRequest 创建博客请求
type CreateBlogRequest struct {
	Title      string `json:"title" binding:"required"`
	Content    string `json:"content" binding:"required"`
	Summary    string `json:"summary"`
	CoverImage string `json:"cover_image"`
	Tags       string `json:"tags"`
	Status     string `json:"status"`
	AuthorID   uint   `json:"-"`
	AuthorName string `json:"-"`
}

// UpdateBlogRequest 更新博客请求
type UpdateBlogRequest struct {
	Title      string `json:"title"`
	Content    string `json:"content"`
	Summary    string `json:"summary"`
	CoverImage string `json:"cover_image"`
	Tags       string `json:"tags"`
	Status     string `json:"status"`
}

// Create 创建博客
func (s *BlogService) Create(req *CreateBlogRequest) (*model.Blog, error) {
	slug := generateSlug(req.Title)
	var count int64
	s.db.Model(&model.Blog{}).Where("slug = ?", slug).Count(&count)
	if count > 0 {
		slug = fmt.Sprintf("%s-%d", slug, time.Now().UnixMilli()%100000)
	}

	status := model.BlogStatusDraft
	if req.Status == string(model.BlogStatusPublished) {
		status = model.BlogStatusPublished
	}

	blog := model.Blog{
		Title:      req.Title,
		Slug:       slug,
		Content:    req.Content,
		Summary:    req.Summary,
		CoverImage: req.CoverImage,
		Tags:       normalizeTags(req.Tags),
		Status:     status,
		AuthorID:   req.AuthorID,
		AuthorName: req.AuthorName,
	}

	if status == model.BlogStatusPublished {
		now := time.Now()
		blog.PublishedAt = &now
	}

	if err := s.db.Create(&blog).Error; err != nil {
		return nil, err
	}
	return &blog, nil
}

// Update 更新博客
func (s *BlogService) Update(id uint, req *UpdateBlogRequest) (*model.Blog, error) {
	var blog model.Blog
	if err := s.db.First(&blog, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("博客不存在")
		}
		return nil, err
	}

	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = req.Title
		slug := generateSlug(req.Title)
		var c int64
		s.db.Model(&model.Blog{}).Where("slug = ? AND id != ?", slug, id).Count(&c)
		if c > 0 {
			slug = fmt.Sprintf("%s-%d", slug, time.Now().UnixMilli()%100000)
		}
		updates["slug"] = slug
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.Summary != "" {
		updates["summary"] = req.Summary
	}
	if req.CoverImage != "" {
		updates["cover_image"] = req.CoverImage
	}
	if req.Tags != "" {
		updates["tags"] = normalizeTags(req.Tags)
	}
	if req.Status != "" {
		newStatus := model.BlogStatus(req.Status)
		if newStatus == model.BlogStatusPublished && blog.Status != model.BlogStatusPublished {
			now := time.Now()
			updates["published_at"] = &now
		}
		updates["status"] = newStatus
	}

	if err := s.db.Model(&blog).Updates(updates).Error; err != nil {
		return nil, err
	}

	s.db.First(&blog, id)
	return &blog, nil
}

// Delete 删除博客（软删除）
func (s *BlogService) Delete(id uint) error {
	result := s.db.Delete(&model.Blog{}, id)
	if result.RowsAffected == 0 {
		return errors.New("博客不存在")
	}
	return nil
}

// GetByID 按 ID 获取博客
func (s *BlogService) GetByID(id uint) (*model.Blog, error) {
	var blog model.Blog
	if err := s.db.First(&blog, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("博客不存在")
		}
		return nil, err
	}
	return &blog, nil
}

// GetBySlug 按 slug 获取已发布博客，并可选增加浏览量
func (s *BlogService) GetBySlug(slug string, incrementView bool) (*model.Blog, error) {
	var blog model.Blog
	if err := s.db.Where("slug = ? AND status = ?", slug, model.BlogStatusPublished).First(&blog).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("博客不存在或未发布")
		}
		return nil, err
	}
	if incrementView {
		s.db.Model(&blog).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1))
		blog.ViewCount++
	}
	return &blog, nil
}

// ListPublished 分页查询已发布博客（公开接口）
func (s *BlogService) ListPublished(page, pageSize int, tag string) ([]model.Blog, int64, error) {
	var blogs []model.Blog
	var total int64

	query := s.db.Where("status = ?", model.BlogStatusPublished)
	if tag != "" {
		query = query.Where("tags LIKE ?", "%"+tag+"%")
	}

	query.Model(&model.Blog{}).Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("published_at DESC").Offset(offset).Limit(pageSize).Find(&blogs).Error; err != nil {
		return nil, 0, err
	}
	return blogs, total, nil
}

// ListAll 分页查询所有博客（管理员使用）
func (s *BlogService) ListAll(page, pageSize int, status, keyword string) ([]model.Blog, int64, error) {
	var blogs []model.Blog
	var total int64

	query := s.db.Model(&model.Blog{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("title LIKE ? OR summary LIKE ?", like, like)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&blogs).Error; err != nil {
		return nil, 0, err
	}
	return blogs, total, nil
}

// GetAllTags 获取所有已用标签
func (s *BlogService) GetAllTags() ([]string, error) {
	rows, err := s.db.Model(&model.Blog{}).
		Where("status = ? AND tags != ''", model.BlogStatusPublished).
		Select("tags").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tagSet := make(map[string]struct{})
	for rows.Next() {
		var t string
		rows.Scan(&t)
		for _, tag := range strings.Split(t, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tagSet[tag] = struct{}{}
			}
		}
	}

	var tags []string
	for t := range tagSet {
		tags = append(tags, t)
	}
	return tags, nil
}

func normalizeTags(tags string) string {
	parts := strings.Split(tags, ",")
	seen := make(map[string]bool)
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" && !seen[p] {
			seen[p] = true
			result = append(result, p)
		}
	}
	return strings.Join(result, ",")
}

func generateSlug(title string) string {
	slug := strings.ToLower(title)
	var b strings.Builder
	prevDash := false
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
			prevDash = r == '-'
		} else if r == ' ' || !prevDash {
			if !prevDash {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	result := strings.Trim(b.String(), "-")
	if result == "" {
		result = fmt.Sprintf("blog-%d", time.Now().UnixMilli())
	}
	return result
}
