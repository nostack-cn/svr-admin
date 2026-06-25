package handler

import "github.com/gin-gonic/gin"

// parseUintParam 解析路径参数中的 uint，非法返回 0
func parseUintParam(c *gin.Context, key string) uint {
	var id uint
	v := c.Param(key)
	if v == "" {
		return 0
	}
	for _, ch := range v {
		if ch < '0' || ch > '9' {
			return 0
		}
		id = id*10 + uint(ch-'0')
	}
	return id
}
