package pagination

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// DefaultPage 默认页码
const DefaultPage = 1

// DefaultPageSize 默认每页条数
const DefaultPageSize = 20

// MaxPageSize 最大每页条数
const MaxPageSize = 100

// Params 分页参数
type Params struct {
	Page     int
	PageSize int
}

// Offset 计算偏移量
func (p Params) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// Limit 返回每页条数
func (p Params) Limit() int {
	return p.PageSize
}

// Parse 从 gin.Context 解析分页参数
func Parse(c *gin.Context) Params {
	page, _ := strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(DefaultPage)))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", strconv.Itoa(DefaultPageSize)))

	if page < 1 {
		page = DefaultPage
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return Params{Page: page, PageSize: pageSize}
}
