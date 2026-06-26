package model

import "time"

// AnnouncementType 公告类型
type AnnouncementType string

const (
	// AnnouncementTypeSite 用于 web-site 首页展示
	AnnouncementTypeSite AnnouncementType = "site"
	// AnnouncementTypeConsole 用于 web-console 控制台展示
	AnnouncementTypeConsole AnnouncementType = "console"
)

// AnnouncementStatus 公告状态
type AnnouncementStatus string

const (
	AnnouncementStatusDraft     AnnouncementStatus = "draft"
	AnnouncementStatusPublished AnnouncementStatus = "published"
)

// Announcement 公告
type Announcement struct {
	BaseModel
	Type        AnnouncementType   `gorm:"type:varchar(20);not null;index" json:"type"`
	Title       string             `gorm:"type:varchar(255);not null" json:"title"`
	Content     string             `gorm:"type:text;not null" json:"content"`
	Status      AnnouncementStatus `gorm:"type:varchar(20);default:'draft';index" json:"status"`
	Priority    int                `gorm:"default:0;index" json:"priority"` // 越大越靠前
	StartsAt    *time.Time         `gorm:"index" json:"starts_at,omitempty"`
	EndsAt      *time.Time         `gorm:"index" json:"ends_at,omitempty"`
	AuthorID    uint               `gorm:"index;not null" json:"author_id"`
	AuthorName  string             `gorm:"type:varchar(64);size:64" json:"author_name"`
	PublishedAt *time.Time         `json:"published_at,omitempty"`
}

func (Announcement) TableName() string {
	return "announcements"
}
