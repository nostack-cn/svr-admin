package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/nostack-cn/svr-admin/model"
)

type blogPost struct {
	Title   string
	Summary string
	Slug    string
	Tags    string
	File    string // 指向 seed/posts/ 下的 Markdown 文件
}

var blogPosts = []blogPost{
	{
		Title:   "域名购买与配置完全指南",
		Summary: "从域名注册、DNS配置到域名解析，一站搞定域名管理全流程。",
		Slug:    "domain-purchase-config-guide",
		Tags:    "运维指南,安全实践",
		File:    "posts/domain-purchase-config-guide.md",
	},
	{
		Title:   "SSL 证书详解：从申请到部署",
		Summary: "深入浅出讲解 SSL/TLS 原理、证书申请方式和 Nginx 部署配置。",
		Slug:    "ssl-certificate-guide",
		Tags:    "安全实践,运维指南",
		File:    "posts/ssl-certificate-guide.md",
	},
	{
		Title:   "域名解析与 CDN 加速最佳实践",
		Summary: "详解域名解析策略、CDN 加速原理及两者的协同配置，提升网站访问速度与稳定性。",
		Slug:    "dns-cdn-best-practices",
		Tags:    "运维指南,技术分享",
		File:    "posts/dns-cdn-best-practices.md",
	},
}

func main() {
	// 数据库连接
	user := envOr("DB_USER", "root")
	password := envOr("DB_PASSWORD", "")
	host := envOr("DB_HOST", "127.0.0.1")
	port := envOr("DB_PORT", "3306")
	dbname := envOr("DB_NAME", "svr_admin")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, dbname)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 获取脚本所在目录（seed/）
	baseDir, _ := os.Getwd()

	now := time.Now()

	for i, p := range blogPosts {
		// 检查 slug 是否已存在
		var count int64
		db.Model(&model.Blog{}).Where("slug = ?", p.Slug).Count(&count)
		if count > 0 {
			log.Printf("[%d/%d] 跳过「%s」— slug=%s 已存在", i+1, len(blogPosts), p.Title, p.Slug)
			continue
		}

		// 读取 Markdown 文件
		contentPath := filepath.Join(baseDir, p.File)
		content, err := os.ReadFile(contentPath)
		if err != nil {
			log.Printf("[%d/%d] 读取文件失败「%s」: %v", i+1, len(blogPosts), p.File, err)
			continue
		}

		blog := model.Blog{
			Title:       p.Title,
			Slug:        p.Slug,
			Content:     string(content),
			Summary:     p.Summary,
			Tags:        p.Tags,
			Status:      model.BlogStatusPublished,
			AuthorID:    1,
			AuthorName:  "nostack",
			PublishedAt: &now,
		}

		if err := db.Create(&blog).Error; err != nil {
			log.Printf("[%d/%d] 创建失败「%s」: %v", i+1, len(blogPosts), p.Title, err)
			continue
		}

		log.Printf("[%d/%d] 已创建「%s」slug=%s", i+1, len(blogPosts), p.Title, p.Slug)
	}

	log.Println("博客种子数据填充完成")
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
