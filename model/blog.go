package model

import "time"

// BlogStatus 博客状态
type BlogStatus string

const (
	BlogStatusDraft     BlogStatus = "draft"
	BlogStatusPublished BlogStatus = "published"
)

// Blog 博客模型
type Blog struct {
	BaseModel
	Title       string     `gorm:"type:varchar(255);not null" json:"title"`
	Slug        string     `gorm:"type:varchar(255);uniqueIndex;size:255;not null" json:"slug"`
	Content     string     `gorm:"type:longtext" json:"content"`
	Summary     string     `gorm:"type:varchar(500);size:500" json:"summary"`
	CoverImage  string     `gorm:"type:varchar(500);size:500" json:"cover_image"`
	Tags        string     `gorm:"type:varchar(500);size:500" json:"tags"`
	Status      BlogStatus `gorm:"type:varchar(20);default:'draft';index" json:"status"`
	AuthorID    uint       `gorm:"index;not null" json:"author_id"`
	AuthorName  string     `gorm:"type:varchar(64);size:64" json:"author_name"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	ViewCount   int        `gorm:"default:0" json:"view_count"`
}

func (Blog) TableName() string {
	return "blogs"
}
