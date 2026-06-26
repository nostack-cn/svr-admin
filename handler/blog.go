package handler

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/nostack-cn/svr-admin/middleware"
	"github.com/nostack-cn/svr-admin/model"
	"github.com/nostack-cn/svr-admin/pkg/pagination"
	"github.com/nostack-cn/svr-admin/pkg/response"
	"github.com/nostack-cn/svr-admin/pkg/upload"
	"github.com/nostack-cn/svr-admin/service"
)

// BlogHandler 博客处理器
type BlogHandler struct {
	blogService *service.BlogService
	uploader    *upload.Uploader
}

// NewBlogHandler 创建博客处理器
func NewBlogHandler(bs *service.BlogService, uploader *upload.Uploader) *BlogHandler {
	return &BlogHandler{blogService: bs, uploader: uploader}
}

// ======================== 管理接口（需 JWT + 权限） ========================

// List GET /api/v1/admin/blogs
func (h *BlogHandler) List(c *gin.Context) {
	p := pagination.Parse(c)
	status := c.Query("status")
	keyword := c.Query("keyword")
	blogs, total, err := h.blogService.ListAll(p.Page, p.PageSize, status, keyword)
	if err != nil {
		response.InternalError(c, "查询博客列表失败")
		return
	}
	response.OKPage(c, blogs, total, p.Page, p.PageSize)
}

// Get GET /api/v1/admin/blogs/:id
func (h *BlogHandler) Get(c *gin.Context) {
	id := parseUint(c, "id")
	if id == 0 {
		response.BadRequest(c, "无效的博客 ID")
		return
	}
	blog, err := h.blogService.GetByID(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, blog)
}

// Create POST /api/v1/admin/blogs
func (h *BlogHandler) Create(c *gin.Context) {
	var req service.CreateBlogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	req.AuthorID = middleware.GetAdminID(c)
	req.AuthorName = middleware.GetAdminUsername(c)

	blog, err := h.blogService.Create(&req)
	if err != nil {
		response.BusinessError(c, 50001, err.Error())
		return
	}
	c.Set("audit_resource", "blog:"+formatUint(blog.ID))
	c.Set("audit_detail", blog.Title)
	response.OK(c, blog)
}

// Update PUT /api/v1/admin/blogs/:id
func (h *BlogHandler) Update(c *gin.Context) {
	id := parseUint(c, "id")
	if id == 0 {
		response.BadRequest(c, "无效的博客 ID")
		return
	}
	var req service.UpdateBlogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	blog, err := h.blogService.Update(id, &req)
	if err != nil {
		response.BusinessError(c, 50002, err.Error())
		return
	}
	c.Set("audit_resource", "blog:"+formatUint(id))
	c.Set("audit_detail", blog.Title)
	response.OK(c, blog)
}

// Delete DELETE /api/v1/admin/blogs/:id
func (h *BlogHandler) Delete(c *gin.Context) {
	id := parseUint(c, "id")
	if id == 0 {
		response.BadRequest(c, "无效的博客 ID")
		return
	}
	if err := h.blogService.Delete(id); err != nil {
		response.BusinessError(c, 50003, err.Error())
		return
	}
	c.Set("audit_resource", "blog:"+formatUint(id))
	response.OKMsg(c, "已删除", nil)
}

// Upload POST /api/v1/admin/blogs/upload
func (h *BlogHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择文件")
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
		".webp": true, ".svg": true, ".bmp": true,
		".pdf": true, ".doc": true, ".docx": true,
		".xls": true, ".xlsx": true, ".ppt": true, ".pptx": true,
		".zip": true, ".rar": true, ".7z": true, ".gz": true,
		".mp4": true, ".mp3": true, ".wav": true,
		".txt": true, ".csv": true, ".json": true, ".xml": true, ".md": true,
	}
	if !allowed[ext] {
		response.BadRequest(c, "不支持的文件类型: "+ext)
		return
	}

	info, err := h.uploader.Save(file, "blogs")
	if err != nil {
		response.BusinessError(c, 50004, err.Error())
		return
	}
	response.OK(c, info)
}

// ServeAdminPage 博客管理后台 HTML 页面
func (h *BlogHandler) ServeAdminPage(c *gin.Context) {
	c.HTML(http.StatusOK, "blog_admin.html", nil)
}

// InternalListBlogs 内部接口 GET /internal/blogs
func (h *BlogHandler) InternalListBlogs(c *gin.Context) {
	p := pagination.Parse(c)
	status := c.Query("status")
	if status == "" {
		status = string(model.BlogStatusPublished)
	}
	blogs, total, err := h.blogService.ListAll(p.Page, p.PageSize, status, "")
	if err != nil {
		response.InternalError(c, "查询博客列表失败")
		return
	}
	response.OKPage(c, blogs, total, p.Page, p.PageSize)
}

// InternalGetBlog 内部接口 GET /internal/blogs/:id
func (h *BlogHandler) InternalGetBlog(c *gin.Context) {
	id := parseUint(c, "id")
	if id == 0 {
		response.BadRequest(c, "无效的博客 ID")
		return
	}
	blog, err := h.blogService.GetByID(id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, blog)
}

// InternalGetBlogBySlug 内部接口 GET /internal/blogs/slug/:slug
func (h *BlogHandler) InternalGetBlogBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		response.BadRequest(c, "无效的博客标识")
		return
	}
	blog, err := h.blogService.GetBySlug(slug, false)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.OK(c, blog)
}

// InternalListTags 内部接口 GET /internal/blogs/tags
func (h *BlogHandler) InternalListTags(c *gin.Context) {
	tags, err := h.blogService.GetAllTags()
	if err != nil {
		response.InternalError(c, "查询标签失败")
		return
	}
	if tags == nil {
		tags = []string{}
	}
	response.OK(c, tags)
}

func parseUint(c *gin.Context, key string) uint {
	var id uint
	if v := c.Param(key); v != "" {
		for _, ch := range v {
			if ch < '0' || ch > '9' {
				return 0
			}
		}
		for _, ch := range v {
			id = id*10 + uint(ch-'0')
		}
	}
	return id
}

func formatUint(n uint) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte(n%10) + '0'
		n /= 10
	}
	return string(buf[i:])
}
